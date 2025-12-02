'''
Disturbance Tracker - Machine Learning Model Definition
'''
import logging

# 3rd-Party
import numpy as np
import torch
from torch import nn
import librosa

# DTrack
import ai.options


# NOTE: Copied from src/ffmpeg/ffmpeg.go
BYTES_PER_SECOND = 96000
SAMPLE_RATE = 48000

# NOTE: Copied from src/model/model.go
SEGMENT_SIZE = 2
SAMPLE_SIZE = SAMPLE_RATE * SEGMENT_SIZE
N_MELS = 128
N_FFT = 2048
HOP_LENGTH = 512

# Mel Spectrogram Filter Bank
MEL_BASIS = librosa.filters.mel(
        htk=True,  # Force HTK math (Matches Go "2595/700" logic)
        sr=SAMPLE_RATE,
        n_fft=N_FFT,
        n_mels=N_MELS,
        fmax=float(SAMPLE_RATE / 2),
        fmin=0.0)

# Use CUDA device if available, or else CPU
CUDA_CPU = torch.device('cuda' if torch.cuda.is_available() else 'cpu')


class NoiseDetector(nn.Module):
    '''
    A PyTorch-based neural network model for detecting specific audio events.
    Uses a simple CNN architecture to ensure ONNX/gonnx compatibility.
    '''
    def __init__(self, num_classes=2):
        super().__init__()

        # Determine output features based on class count
        # For Binary (1 output), Multi-class (N outputs)
        # Note: We use CrossEntropyLoss for multi-class, so we need N outputs.
        # If num_classes is 2, we technically could use 1 with Sigmoid, but
        # we use "N outputs + Softmax" to support arbitrary number of classes
        self.num_classes = num_classes

        self.features = nn.Sequential(
            # Block 1: Output (64, 64, 94) - (1/2 size)
            nn.Conv2d(1, 32, kernel_size=3, padding=1),
            nn.BatchNorm2d(32),
            nn.ReLU(inplace=True),
            nn.Conv2d(32, 64, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(64),
            nn.ReLU(inplace=True),

            # Block 2: Output (128, 32, 47) - (1/4 size)
            nn.Conv2d(64, 128, kernel_size=3, padding=1),
            nn.BatchNorm2d(128),
            nn.ReLU(inplace=True),
            nn.Conv2d(128, 128, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(128),
            nn.ReLU(inplace=True),

            # Block 3: Output (256, 16, 23) - (1/8 size)
            nn.Conv2d(128, 256, kernel_size=3, padding=1),
            nn.BatchNorm2d(256),
            nn.ReLU(inplace=True),
            nn.Conv2d(256, 256, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(256),
            nn.ReLU(inplace=True),

            # Block 4: Output (512, 8, 12) - (1/16 size)
            nn.Conv2d(256, 512, kernel_size=3, padding=1),
            nn.BatchNorm2d(512),
            nn.ReLU(inplace=True),
            nn.Conv2d(512, 512, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(512),
            nn.ReLU(inplace=True),

            # Block 5: Output (512, 4, 6) - (1/32 size)
            nn.Conv2d(512, 512, kernel_size=3, padding=1),
            nn.BatchNorm2d(512),
            nn.ReLU(inplace=True),
            nn.Conv2d(512, 512, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(512),
            nn.ReLU(inplace=True),

            # Block 6: (Maximum Depth) Output (1024, 2, 3) - (1/64 size)
            nn.Conv2d(512, 1024, kernel_size=3, padding=1),
            nn.BatchNorm2d(1024),
            nn.ReLU(inplace=True),
            nn.Conv2d(1024, 1024, kernel_size=3, padding=1, stride=2),
            nn.BatchNorm2d(1024),
            nn.ReLU(inplace=True),
        )

        self.classifier = nn.Sequential(
            nn.AdaptiveMaxPool2d(1),
            nn.Flatten(),
            nn.Dropout(p=0.3),
            nn.Linear(1024, num_classes)
        )

    def forward(self, x):
        '''
        Defines the forward pass of the model.
        '''
        x = self.features(x)
        return self.classifier(x)


def save(model, path):
    '''
    Saves the model state to a .pth file.
    '''
    path.parent.mkdir(parents=True, exist_ok=True)
    logging.trace('Saving model to %s', path)
    torch.save(model.state_dict(), path)


def load(path, num_classes):
    '''
    Returns a a trained model from a .pth file as NoiseDetector object.
    Must provide num_classes to initialize the architecture correctly.
    '''
    logging.trace('Loading model from %s (Classes: %d)', path, num_classes)
    device = ai.model.CUDA_CPU
    model = NoiseDetector(num_classes=num_classes)
    model.load_state_dict(torch.load(path, map_location=device))
    model.to(device)
    model.eval()  # Set model to evaluation mode
    return model


def convert(pth, onnx, num_classes):
    '''
    Convert pytorch .pth model to ONNX (open model) format.
    '''
    logging.info('Converting %s to %s', pth, onnx)
    model = load(pth, num_classes)
    torch.onnx.export(
        model, torch.randn(1, 1, 128, 188).to(ai.model.CUDA_CPU), onnx,
        input_names=['input'], output_names=['output'], opset_version=20,
        dynamo=False, verbose=False)


def open_audio_file(filepath):
    '''
    Loads raw PCM audio data from a client-formatted .dat file.
    '''
    with open(filepath, 'rb') as fh:
        return normalize_audio(fh.read())


def normalize_audio(pcm_data):
    '''
    Convert 2-second buffer to numpy array and normalize.
    '''
    audio_data = np.frombuffer(pcm_data, dtype=np.int16)
    return audio_data.astype(np.float32) / 32768.0


def pad_audio(raw_data):
    '''
    Ensure audio slice is an exact size. (should not be needed)
    '''
    # Pad with silence (zeros) if the clip is too short
    if len(raw_data) < SAMPLE_SIZE:
        raw_data += b'\x00' * (SAMPLE_SIZE - len(raw_data))
    # Truncate if the clip is too long
    elif len(raw_data) > SAMPLE_SIZE:
        raw_data = raw_data[:SAMPLE_SIZE]
    return raw_data


def audio_to_spectrogram(audio_data):
    '''
    Converts to Power Spectrogram and RESIZES to match
    the model's expected N_MELS dimension.
    '''
    # Manually Compute Power STFT
    stft = np.abs(librosa.stft(
        y=audio_data,
        n_fft=N_FFT,
        hop_length=HOP_LENGTH)) ** 2

    # Apply cached Mel Lens (Matrix Multiplication)
    s_mel = np.dot(MEL_BASIS, stft)
    s_db = librosa.power_to_db(s_mel, ref=np.max)

    # Fixed Normalization (-80dB floor)
    s_db = np.clip(s_db, -80, 0)
    img = (s_db + 80) / 80

    # Return single-dimension channel (N, 1, H, W)
    return torch.tensor(np.expand_dims(img, axis=0), dtype=torch.float32)
