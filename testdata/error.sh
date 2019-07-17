#!/bin/bash

echo "this script fails with error"
>&2 echo "prerequisites not found"
exit 3
