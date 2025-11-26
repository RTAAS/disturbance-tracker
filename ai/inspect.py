'''
Disturbance Tracker - Inspection Utility (Multi-Class)
'''
import logging
import pathlib
import pprint
import sys
import json
import torch
import torch.nn.functional as F

# DTrack project imports
import ai.options
import ai.model


def check_input():
    '''
    Main entry point for the inspection script.
    '''
    options = ai.options.bootstrap()
    workspace = pathlib.Path(options['workspace'])

    # Load models and their labels
    loaded_models = {}

    if not options['inspect_models']:
        raise ValueError('No inspection models are configured.')

    for model_name in options['inspect_models']:
        # Load Labels
        labels_path = workspace / 'models' / f'{model_name}_labels.json'
        if not labels_path.exists():
            raise FileNotFoundError(f'Labels missing for {model_name}')

        with open(labels_path, 'r') as fh:
            labels = json.load(fh)

        # Load Model
        pth_path = workspace / 'models' / f'{model_name}.pth'
        model = ai.model.load(pth_path, num_classes=len(labels))

        loaded_models[model_name] = {
            'model': model,
            'labels': labels}

    # Execution Routing
    if not options.get('inspect_path'):
        logging.debug('Running inference with standard input')
        audio = sys.stdin.buffer.read()
        if not audio:
            raise ValueError('No standard input!')
        pprint.pprint(infer_all(loaded_models, audio))
    else:
        inspect_path = pathlib.Path(options['inspect_path'])
        if inspect_path.is_file():
            logging.debug('Running inference with single file')
            for audio in slice_audio(inspect_path):
                pprint.pprint(infer_all(loaded_models, audio))

        elif inspect_path.is_dir():
            logging.debug('Running inference with directory of mkv files')
            for filename in inspect_path.glob('*.*'):
                logging.info('Reviewing %s', filename)
                for audio in slice_audio(filename):
                    pprint.pprint(infer_all(loaded_models, audio))
        else:
            raise OSError(f'Could not find {inspect_path}')


def slice_audio(path):
    '''
    Return a list of numpy-prepared 2-second segments from audio file
    '''
    # Test against tagged pcm data files
    if path.match('*.dat'):
        return [ai.model.open_audio_file(path)]

    logging.warning('Skipping %s (wrong file type or not implemented)', path)
    return []


def infer_all(model_bundle, audio_data):
    '''
    Run inference on single audio segment using all trained models.
    Expects model_bundle = {'name': {'model': m, 'labels': [...]}}
    '''
    # Convert raw pcm bytes to numpy array
    if isinstance(audio_data, bytes):
        audio_data = ai.model.normalize_audio(audio_data)

    # Convert the NumPy array into a spectrogram tensor.
    spectrogram = ai.model.audio_to_spectrogram(audio_data)

    # Add a batch dimension (B, C, H, W) for the model
    input_tensor = spectrogram.unsqueeze(0).to(ai.model.CUDA_CPU)

    # Inference Step
    results = {}
    with torch.no_grad():
        for name, data in model_bundle.items():
            model = data['model']
            labels = data['labels']

            # Forward pass (Logits)
            logits = model(input_tensor)

            # Softmax to get probabilities (sum to 1.0)
            probs = F.softmax(logits, dim=1).squeeze().tolist()

            # Handle single class case or list conversion
            if not isinstance(probs, list):
                probs = [probs]

            # Create readable dictionary
            # e.g., {'empty': 0.1, 'barking': 0.9}
            class_probs = {labels[i]: round(probs[i], 4)
                           for i in range(len(labels))}

            # Get best match
            best_idx = torch.argmax(logits, dim=1).item()
            best_label = labels[best_idx]

            results[name] = {
                'match': best_label,
                'confidence': class_probs[best_label],
                'distribution': class_probs
            }

    return results


if __name__ == '__main__':
    check_input()
