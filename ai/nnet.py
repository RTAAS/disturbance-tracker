'''
PyTorch Model
'''
import logging
import pathlib
import glob
import shutil
import torch
import torchaudio

# DTrack
import ai.options
import ai.nnet


class AudioClassifier:
    '''
    TODO
    '''
    def __init__(self, model_name):
        # Check for cuda
        self.device = 'cuda' if torch.cuda.is_available() else 'cpu'
        logging.debug(f'Backend: {self.device}')

        # Data locations
        self.workspace = pathlib.Path(ai.options.get('workspace'))
        self.model_path = self.workspace / f'{model_name}.pth'
        self.training_data = self.workspace / 'train'
        self.testing_data = self.workspace / 'test'
        self.sample_rate = None
        self._primed = False
        self._loaders = {}

        # Tuning options
        self.batch_size = 1
        self.learn_rate = ai.options.get('learning_rate')

        # Models (search labels)
        self.models = ai.options.get('inspect_models')
        self.label2index = {m: i for i, m in enumerate(self.models)}
        self.index2label = {i: m for i, m in enumerate(self.models)}

        # ML Model - Assumes single channel audio
        self.network = M5(n_input=1).to(self.device)
        if self.model_path.exists():
            logging.debug(f'Loading previous state from {self.model_path}')
            self.network.load_state_dict(torch.load(self.model_path))
            self._primed = True
        logging.debug(f'Number of params: {count_parameters(self.network)}')

        # Training state
        self.optimizer = torch.optim.SGD(
                self.network.parameters(),
                lr=self.learn_rate,
                momentum=ai.options.get('train_momentum'))
        self.scheduler = torch.optim.lr_scheduler.MultiStepLR(
                self.optimizer, milestones=[10, 30], gamma=0.1)

        # Load a sample clip to build transformation
        self._load_sample()

    def _load_sample(self):
        '''
        Sample the first available input clip
        '''
        sample_clip = self.workspace / 'model.wav'

        # Grab the first available clip if sample does not exist
        if not sample_clip.exists():
            demo_clip = next((self.training_data / 'nomatch').glob('*.wav'))
            shutil.copy(demo_clip, sample_clip)
            logging.debug(f'Generated model.wav sample from {demo_clip}')

        # Load sample for the transformation engine
        logging.debug(f'Generated model.wav sample from {sample_clip}')
        self.waveform, self.sample_rate = self.load_audio(sample_clip)

    def load_audio(self, file_path):
        '''
        Load audio and transform to mono channel
        '''
        wv, sample_rate = torchaudio.load(file_path)
        # Convert multiple channels to mono
        if wv.shape[0] > 1:
            wv = wv.mean(dim=0, keepdim=True)
        return (wv, sample_rate)

    def get_loader(self, data):
        '''
        Returns a data loader after ensuring it is loaded
        '''
        # Return loader if already loaded
        if data in self._loaders:
            return self._loaders[data]

        # Load specified loader
        path = getattr(self, data)
        dataset = NoiseDataset(
                root_dir=path, models=self.models)
        loader = torch.utils.data.DataLoader(
                dataset, batch_size=self.batch_size,
                shuffle=True, collate_fn=collate_fn)

        # Save and return loader
        self._loaders[data] = loader
        return loader

    def training_loop(self):
        '''
        Continue testing new models until target_accuracy is met
        '''
        # Track best iteration
        best_i = 0
        # Use defaults if no model was loaded
        if not self._primed:
            best = dict.fromkeys(self.models, 0)
            avg_best = 0
        else:
            best = self.evaluate()
            avg_best = sum(best.values()) / len(best)
            logging.info(f'Initial accuracy[0] is {avg_best}')

        # Continue training until desired threshold is met
        iteration = 0
        last_accuracy = [0, 0]  # [accuracy, count]
        while avg_best < float(ai.options.get('train_target')):
            iteration += 1

            # Train a new model
            logging.debug(f'Training iteration {iteration}')
            self.train_once()

            # Check accuracy of new model
            logging.debug(f'Testing accuracy of iteration {iteration}')
            accuracy = self.evaluate()
            logging.info('Overall accuracy[{i}] is {a}'.format(
                i=iteration, a=sum(accuracy.values()) / len(best)))

            # Save the model if accuracy improves
            logging.debug('Old Accuracy: {old}  New Accuracy: {new}'.format(
                old=sum(best.values()), new=sum(accuracy.values())))
            if sum(accuracy.values()) > sum(best.values()):
                logging.info('Accuracy increased; keeping new model')
                torch.save(self.network.state_dict(), self.model_path)
                best = accuracy.copy()
                best_i = iteration
                avg_best = sum(best.values()) / len(best)
            else:
                logging.info(f'Accuracy worse than #{best_i}; discarding new')

            # Check for an infinite loop (if target_accuracy cannot be met)
            if sum(accuracy.values()) != last_accuracy[0]:
                last_accuracy = [sum(accuracy.values()), 0]
            else:
                last_accuracy[1] += 1

                # Attempt to get different results with modified learning rate
                if last_accuracy[1] % 3 == 0:
                    logging.warning(
                            'No accuracy change for 3 rounds; bumping entropy')
                    # Randomly adjust learning rate
                    self.learn_rate *= (1 + (torch.rand(1).item() - 0.5) * 0.1)
                    # Update optimizer with the new learning rate
                    for param_group in self.optimizer.param_groups:
                        param_group['lr'] = self.learn_rate

                # Give up if entropy bump produced no changes
                if last_accuracy[1] >= 10:
                    logging.critical(
                            'No accuracy change for 10 rounds; stopping')
                    # Exit loop (starting fresh with the same model is best)
                    break

        logging.info(f'TRAINING COMPLETE :: Final Accuracy: {avg_best}')

    def train_once(self):
        '''
        Train a model using collected data
        '''
        criterion = torch.nn.CrossEntropyLoss()
        running_loss, correct, total = 0.0, 0, 0
        for i, (inputs, labels) in enumerate(self.get_loader('training_data')):

            # Zero the parameter gradients
            self.optimizer.zero_grad()

            # Forward + backward + optimize
            outputs = self.network(inputs.to(self.device)).squeeze(1)
            loss = criterion(outputs, labels.to(self.device))
            loss.backward()
            self.optimizer.step()

            running_loss += loss.item()
            _, predicted = torch.max(outputs.data, 1)
            total += labels.to(self.device).size(0)
            correct += (predicted == labels.to(self.device)).sum().item()

            # Print statistics every 20 mini-batches
            if i % 2000 == 1999:
                logging.debug('#{i:5d} Loss: {l:.3f} Accuracy: {a:.1f}'.format(
                    i=i, l=running_loss / 2000, a=100 * correct // total))
                # Reset tally
                running_loss, correct, total = 0.0, 0, 0

        self.scheduler.step()

    def evaluate(self):
        correct_pred = dict.fromkeys(self.models, 0)
        total_pred = dict.fromkeys(self.models, 0)

        with torch.no_grad():
            for inputs, labels in self.get_loader('testing_data'):
                outputs = self.network(inputs.to(self.device)).squeeze(1)
                _, predictions = torch.max(outputs, 1)

                # Collect correct predictions for each class
                for label, prediction in zip(
                        labels.to(self.device), predictions):
                    if label == prediction:
                        correct_pred[self.index2label[label.item()]] += 1
                    total_pred[self.index2label[label.item()]] += 1

        # Calculate accuracy for each class
        accuracy = {}
        for cls, correct_count in correct_pred.items():
            accuracy[cls] = 100 * float(correct_count) / total_pred[cls]
            logging.debug(
                    f'Accuracy for {cls:5s} is {accuracy[cls]:.1f}% '
                    f'({correct_count} of {total_pred[cls]})')

        return accuracy


class M5(torch.nn.Module):
    '''
    Convolutional neural network with multiple convolutional and pooling layers
    '''
    def __init__(self, n_input=1, n_output=2, stride=16, n_channel=16):
        super().__init__()
        nn = torch.nn
        self.conv1 = nn.Conv1d(n_input, n_channel, kernel_size=49, stride=16)
        self.bn1 = nn.BatchNorm1d(n_channel)
        self.pool1 = nn.MaxPool1d(4)
        self.conv2 = nn.Conv1d(n_channel, 2 * n_channel, kernel_size=49)
        self.bn2 = nn.BatchNorm1d(2 * n_channel)
        self.pool2 = nn.MaxPool1d(2)
        self.conv3 = nn.Conv1d(2 * n_channel, 4 * n_channel, kernel_size=7)
        self.bn3 = nn.BatchNorm1d(4 * n_channel)
        self.pool3 = nn.MaxPool1d(2)
        self.conv4 = nn.Conv1d(4 * n_channel, 2 * n_channel, kernel_size=5)
        self.bn4 = nn.BatchNorm1d(2 * n_channel)
        self.pool4 = nn.MaxPool1d(2)
        self.conv5 = nn.Conv1d(2 * n_channel, 1 * n_channel, kernel_size=3)
        self.bn5 = nn.BatchNorm1d(1 * n_channel)

        # Adaptive Global Average Pool (GAP)
        self.global_avg_pool = nn.AdaptiveAvgPool1d(1)
        # Prevent over-fitting
        self.dropout = nn.Dropout(ai.options.get('train_dropout'))
        # Adjusted input size after GAP
        self.fc1 = nn.Linear(n_channel, n_output)

    def forward(self, x):
        F = torch.nn.functional
        x = self.conv1(x)
        x = F.relu(self.bn1(x))
        x = self.pool1(x)
        x = self.conv2(x)
        x = F.relu(self.bn2(x))
        x = self.pool2(x)
        x = self.conv3(x)
        x = F.relu(self.bn3(x))
        x = self.pool3(x)
        x = self.conv4(x)
        x = F.relu(self.bn4(x))
        x = self.pool4(x)
        x = self.conv5(x)
        x = F.relu(self.bn5(x))

        # Use global average pooling
        x = self.global_avg_pool(x)
        # Flatten for fully connected layer
        x = x.view(x.size(0), -1)
        # Apply dropout
        x = self.dropout(x)

        x = self.fc1(x)
        return x


class NoiseDataset(torch.utils.data.Dataset):
    def __init__(self, root_dir, models, transform=[], n_input=256):
        self.audio_list = glob.glob(f'{root_dir}/*/*.wav')
        self.transform = transform
        self.n_input = n_input
        self.mean, self.std = self._compute_mean()
        self.label2index = {m: i for i, m in enumerate(models)}

    def _compute_mean(self):
        meanstd_file = pathlib.Path(ai.options.get('workspace')) / '_mean.pth'
        if meanstd_file.exists():
            meanstd = torch.load(meanstd_file)
        else:
            logging.debug('computing _mean.pth')
            mean = torch.zeros(self.n_input)
            std = torch.zeros(self.n_input)
            cnt = 0
            for path in self.audio_list:
                cnt += 1
                logging.debug(f' {cnt} | {len(self.audio_list)}')

                wv, _ = torchaudio.load(path)
                if wv.shape[0] > 1:
                    wv = torch.mean(wv, axis=0, keepdim=True)

                for tr in self.transform:
                    wv = tr(wv)
                mean += wv.mean(1)
                std += wv.std(1)

            mean /= len(self.audio_list)
            std /= len(self.audio_list)
            meanstd = {
                'mean': mean,
                'std': std,
                }
            torch.save(meanstd, meanstd_file)

        return meanstd['mean'], meanstd['std']

    def __len__(self):
        return len(self.audio_list)

    def __getitem__(self, idx):
        path = self.audio_list[idx]
        class_name = path.split('/')[-2]
        label = self.label2index[class_name]

        wv, sr = torchaudio.load(path)
        if wv.shape[0] > 1:
            wv = torch.mean(wv, axis=0, keepdim=True)
        audio_feature = wv
        if self.transform:
            for tr in self.transform:
                audio_feature = tr(audio_feature)

        return audio_feature, torch.tensor(label)


def collate_fn(batch):
    '''
    Aggregates a list of data samples into a batched tensor
    suitable for input into a neural network
    '''
    # A data tuple has the form: waveform, label
    tensors, targets = [], []

    # Gather in lists, and encode labels as indices
    for waveform, label in batch:
        tensors += [waveform]
        targets += [label]

    # Group the list of tensors into a batched tensor
    tensors = pad_sequence(tensors)
    targets = torch.stack(targets)

    return tensors, targets


def pad_sequence(batch):
    '''
    Pad all tensors in a batch to the same length using zeros
    '''
    batch = [item.t() for item in batch]
    batch = torch.nn.utils.rnn.pad_sequence(
            batch, batch_first=True, padding_value=0.)
    return batch.permute(0, 2, 1)


def count_parameters(model):
    '''
    Return the number of trainable parameters in a PyTorch model
    '''
    return sum(p.numel() for p in model.parameters() if p.requires_grad)
