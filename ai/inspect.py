'''
Disturbance Tracker - Inspection Utility
'''
import pathlib
import tempfile
import torch

# DTrack
import ai.options
import ai.nnet


def main():
    '''
    Perform inference with a specified model against a given input.
    '''
    print(ai.options.get("train_rate"))
    inspect_path = pathlib.Path(ai.options.get('mkv_path'))
    if inspect_path.is_file():
        mframes = scan_mkv(inspect_path)
        print(sorted(mframes))
    elif inspect_path.is_dir():
        for mkv in inspect_path.glob('*.mkv'):
            mframes = scan_mkv(mkv)
            print(f'{mkv.name}: {sorted(mframes)}')
    else:
        raise Exception(f'Could not find {inspect_path}')


def scan_mkv(audio_file):
    '''
    Review each audio segment for a match to the trained model
    '''
    audio_path = pathlib.Path(audio_file)
    if not audio_path.exists():
        raise Exception(f'No video was found: {audio_file}')

    classifier = ai.nnet.AudioClassifier()
    tags = [t for t in ai.options.get('inspect_models') if t != 'nomatch']

    # Extract 1-second clips
    tempdir = tempfile.TemporaryDirectory()
    apr.common.extract_audio(audio_path, tempdir.name)

    # Review each clip
    # TODO: This does not support multiple tags
    matched_frames = []
    for wav in pathlib.Path(tempdir.name).glob('*.wav'):
        transformed = classifier.load_audio(wav)[0]
        inputs = transformed.unsqueeze(0)
        with torch.no_grad():
            output = classifier.network(inputs).squeeze(1)
            _, prediction = torch.max(output, len(tags))

        # Check the classification result
        if classifier.index2label[prediction.item()] in tags:
            # Keep track of matched frames
            matched_frames.append(int(wav.name.split('.')[0]))

    return matched_frames


if __name__ == '__main__':
    main()
