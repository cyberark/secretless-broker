#!/usr/bin/env python3

# Ensure we are on Python3
import sys
assert sys.version_info >= (3, 0), "Python3 should be used to run this program"

import argparse
import re

RUN_PATTERN = re.compile('.* \[\d+\/')
TIME_PATTERN = re.compile("^([\d\.]*)(\w*)$")

CHECKPOINTS = [1.10, 1.15, 1.20, 1.30, 1.50, 1.60, 1.70]
BASELINE_SMOOTHING_POINTS = 50


def process_log(logfile_path, baseline_backend):
    percentages = {}
    baselines = []
    with open(logfile_path) as log_file:
        for line in log_file:
            if not RUN_PATTERN.match(line):
                continue

            line = line.strip()
            fields = line.split()

            backend = re.sub(r"/.*", "", fields[3])
            time = fields[5].replace('s,', 's')
            (numbers, scale) = TIME_PATTERN.search(time).groups()
            time = float(numbers)
            if scale == "ms":
                time *= 1000
            elif scale == "s":
                time *= 1000000

            if backend == baseline_backend:
                baselines.append(time)
                baselines = baselines[-BASELINE_SMOOTHING_POINTS:]
            else:
                if not backend in percentages:
                    percentages[backend] = []

                baseline_average = 0.0
                if baselines:
                    baseline_average = sum(baselines) / len(baselines)

                    percentages[backend].append(time / baseline_average)

    print("Baseline backend:", baseline_backend)
    if not baselines:
        raise RuntimeError("ERROR: Could not find any datapoints for '{}'!".format(baseline_backend))

    for backend in percentages.keys():
        datapoint_count = len(percentages[backend])
        print("Backend:", backend)
        print("Count:", datapoint_count)
        percentages[backend].sort()

        for checkpoint in CHECKPOINTS:
            index_of_next = next(x[0] for x in enumerate(percentages[backend]) if x[1] >= checkpoint)

            datapoints_below_checkpoint = float(index_of_next) / datapoint_count * 100
            print("Below %3.0f%% of baseline: %3.2f%% of requests." % \
                    (checkpoint * 100, datapoints_below_checkpoint))

        print()

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Program use to take aggregate Juxtaposer logs and produce sample bounds',
    )

    parser.add_argument('logfile', action='store', type=str)
    parser.add_argument('baseline_name', action='store', type=str)

    args = parser.parse_args()
    process_log(args.logfile, args.baseline_name)
