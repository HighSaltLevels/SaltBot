#!/bin/bash

PYTHONPATH=$(pwd)/saltbot coverage run --source=saltbot -m pytest tests
if [ $? != 0 ]; then
    echo "One or more tests failed!"
    exit 1
fi

coverage report -m
