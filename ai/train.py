'''
Disturbance Tracker - Model Trainer
'''
import ai.nnet
import ai.options


def main():
    # TODO: Start a parallel process group
    for model in ai.options.get('inspect_models'):
        classifier = ai.nnet.AudioClassifier(model)
        classifier.training_loop()
    # TODO: Wait for processes to complete


if __name__ == '__main__':
    main()
