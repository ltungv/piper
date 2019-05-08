import random
import time
import json
import sys
import os


COLORS = ['k', 'r', 'g', 'b', 'y']
FPS = 30
names = []

def randData():
    frame = []
    for name in names:
        position = [random.randint(0, 1500), random.randint(0, 1500)]
        dimension = [random.randint(-90, 90), random.randint(-90, 90)]
        frame.append({
            'name': name,
            'pos': position,
            'dim': dimension
        })

    return frame


def main():
    for i in range(len(COLORS)):
        for j in range(i, len(COLORS)):
            names.append(COLORS[i] + COLORS[j])

    while True:
        print(json.dumps(randData()))
        sys.stdout.flush()
        time.sleep(1 / FPS)


if __name__ == '__main__':
    main()
