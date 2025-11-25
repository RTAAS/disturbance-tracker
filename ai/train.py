'''
Disturbance Tracker - Model Trainer (Multi-Class with Class Balancing)
'''
import logging
import pathlib
import json
import audiomentations
import sklearn.model_selection
import torch
import tqdm
import sys # Added for stdout flushing

# DTrack
import ai.model
import ai.options


class AudioDataset(torch.utils.data.Dataset):
    '''
    Custom PyTorch Dataset for loading and processing audio files.
    '''
    def __init__(self, file_paths, labels, augmentations=None):
        self.file_paths = file_paths
        self.labels = labels # Expecting Integers (0, 1, 2...)
        self.augmentations = augmentations

    def __len__(self):
        return len(self.file_paths)

    def __getitem__(self, idx):
        filepath = self.file_paths[idx]
        label = self.labels[idx]

        # Load audio using the function from model.py
        audio = ai.model.open_audio_file(filepath)

        # Apply augmentations if specified
        if self.augmentations:
            # Wrap in try-except to handle edge cases in augmentation
            try:
                audio = self.augmentations(
                        samples=audio, sample_rate=ai.model.SAMPLE_RATE)
            except Exception:
                pass

        # Convert audio to a spectrogram
        spectrogram = ai.model.audio_to_spectrogram(audio)

        # Return Label as LongTensor for CrossEntropyLoss
        return spectrogram, torch.tensor(label, dtype=torch.long)


def setup_file_logging(workspace_path):
    '''
    Sets up a file handler to log output to training.log
    '''
    log_file = workspace_path / 'training.log'
    
    # Create file handler
    file_handler = logging.FileHandler(log_file, mode='w')
    file_handler.setLevel(logging.INFO)
    
    # Create formatter
    formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
    file_handler.setFormatter(formatter)
    
    # Add handler
    logging.getLogger().addHandler(file_handler)
    logging.info(f"Logging training results to: {log_file}")


def train_all():
    '''
    Main entry point for the training script.
    '''
    options = ai.options.bootstrap()
    workspace = pathlib.Path(options['workspace'])
    models_dir = workspace / 'models'

    # Setup logging to file
    setup_file_logging(workspace)

    for model_name in options['inspect_models']:
        logging.info('Begin training: %s', model_name)
        try:
            # Step 1: Train and get class count
            num_classes = train_model(model_name, options)
            logging.info('Finished training %s', model_name)
            
            # Step 2: Export to ONNX with correct class count
            ai.model.convert(
                models_dir / f'{model_name}.pth',
                models_dir / f'{model_name}.onnx',
                num_classes
            )
            logging.info('MODEL PREPARED: %s', model_name)
            
        except KeyboardInterrupt:
            logging.info('Received Ctrl+C; Training stopped')
        except Exception as e:
            logging.error('Failed to train %s: %s', model_name, e)
            raise e


def train_model(model_name, options):
    '''
    Conducts the full training and validation process for a model.
    Returns: int (number of classes)
    '''
    workspace = pathlib.Path(options['workspace'])
    models_dir = workspace / 'models'
    tags_dir = workspace / 'tags'
    
    # 1. Dynamic Dataset Discovery
    if not tags_dir.exists():
        raise OSError(f'Tags directory not found: {tags_dir}')

    # Get all subdirectories (classes) and sort them
    class_folders = sorted([d for d in tags_dir.iterdir() if d.is_dir()])
    
    if len(class_folders) < 2:
        raise ValueError(
            f"Need at least 2 class folders in {tags_dir} for classification. Found: {[d.name for d in class_folders]}")

    # Create mapping: Name -> Index
    classes = [d.name for d in class_folders]
    class_to_idx = {cls_name: i for i, cls_name in enumerate(classes)}
    
    logging.info("Detected Classes: %s", class_to_idx)

    # Ensure models directory exists
    models_dir.mkdir(parents=True, exist_ok=True)

    # Save Labels Map
    labels_path = models_dir / f'{model_name}_labels.json'
    with open(labels_path, 'w') as fh:
        json.dump(classes, fh)

    # Gather files and Calculate Counts for Balancing
    all_files = []
    all_labels = []
    class_counts = [] 

    for cls_name in classes:
        folder = tags_dir / cls_name
        files = list(folder.glob('*.dat'))
        count = len(files)
        logging.info("Class '%s': %d samples", cls_name, count)
        class_counts.append(count) # Store count
        
        all_files.extend(files)
        all_labels.extend([class_to_idx[cls_name]] * len(files))

    if not all_files:
        raise OSError('No .dat files found in class directories.')

    # Calculate Class Weights for Imbalance
    # Formula: Total / (NumClasses * ClassCount)
    total_samples = sum(class_counts)
    num_classes = len(classes)
    class_weights = [total_samples / (num_classes * c) for c in class_counts]
    logging.info("Class Weights: %s", class_weights)
    # --------------------------------------------------

    # Stratified Split
    train_files, val_files, train_labels, val_labels = \
        sklearn.model_selection.train_test_split(
                all_files, all_labels, test_size=0.2,
                random_state=42, stratify=all_labels)

    logging.info(
            'Training samples: %d Validation samples: %d',
            len(train_files), len(val_files))

    # 2. Augmentations & Loaders
    train_dataset = AudioDataset(
            train_files,
            train_labels,
            augmentations=audiomentations.Compose([
                audiomentations.AddGaussianNoise(
                    min_amplitude=0.001, max_amplitude=0.015, p=0.5),
                audiomentations.TimeStretch(
                    min_rate=0.8, max_rate=1.25, p=0.5),
                audiomentations.PitchShift(
                    min_semitones=-4, max_semitones=4, p=0.5),
            ]))
    val_dataset = AudioDataset(
            val_files,
            val_labels,
            augmentations=None)

    train_loader = torch.utils.data.DataLoader(
            train_dataset,
            batch_size=options['train_batch_size'],
            shuffle=True)
    val_loader = torch.utils.data.DataLoader(
            val_dataset,
            batch_size=options['train_batch_size'],
            shuffle=False)

    # 3. Initialize Model
    logging.debug('Creating %s model with %d output classes', model_name, num_classes)
    
    dev = ai.model.CUDA_CPU
    model = ai.model.NoiseDetector(num_classes=num_classes).to(dev)
    
    # Apply Weights to Loss Function
    weights_tensor = torch.tensor(class_weights, dtype=torch.float32).to(dev)
    criterion = torch.nn.CrossEntropyLoss(weight=weights_tensor)
    
    optimizer = torch.optim.AdamW(
        model.parameters(), lr=options['train_learning_rate'])
    scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(
        optimizer, 'min', patience=5)

    # 4. Training Loop
    best_val_loss = float('inf')
    epochs_worse = 0
    
    for epoch in range(options['train_epochs']):

        # Training Phase
        model.train()
        train_loss, train_correct = 0, 0
        
        # Use sys.stdout for Tqdm to prevent log file corruption
        for inputs, labels in tqdm.tqdm(
                train_loader, desc=f'Epoch {epoch} [Train]', leave=False, file=sys.stdout):
            inputs = inputs.to(dev)
            labels = labels.to(dev)
            
            optimizer.zero_grad()
            outputs = model(inputs)
            loss = criterion(outputs, labels)
            loss.backward()
            optimizer.step()
            
            train_loss += loss.item()
            
            # Accuracy: Max argument of Softmax
            _, preds = torch.max(outputs, 1)
            train_correct += (preds == labels).sum().item()

        # Validation Phase
        model.eval()
        val_loss, val_correct = 0, 0
        with torch.no_grad():
            for inputs, labels in tqdm.tqdm(
                    val_loader,
                    desc=f'Epoch {epoch} [Check]',
                    leave=False, file=sys.stdout):
                inputs = inputs.to(dev)
                labels = labels.to(dev)
                
                outputs = model(inputs)
                loss = criterion(outputs, labels)
                val_loss += loss.item()
                
                _, preds = torch.max(outputs, 1)
                val_correct += (preds == labels).sum().item()

        # Log Results (Using logging.info to show in file)
        avg_train_loss = train_loss / len(train_loader)
        avg_val_loss = val_loss / len(val_loader)
        train_accuracy = 100 * train_correct / len(train_dataset)
        val_accuracy = 100 * val_correct / len(val_dataset)

        # logging.info(
        #     'Epoch %d: Train Loss: %.4f, Acc: %.2f%% | Val Loss: %.4f, Acc: %.2f%%',
        #     epoch, avg_train_loss, train_accuracy, avg_val_loss, val_accuracy)

        scheduler.step(avg_val_loss)

        if avg_val_loss < best_val_loss:
            logging.info('Model improved (Loss: %.4f). Saving.', avg_val_loss)
            ai.model.save(model, models_dir / f'{model_name}.pth')
            best_val_loss = avg_val_loss
            epochs_worse = 0
        else:
            epochs_worse += 1

        if epochs_worse >= options['train_patience']:
            logging.info('Training patience exhausted without improvement')
            return num_classes

    return num_classes


if __name__ == '__main__':
    train_all()