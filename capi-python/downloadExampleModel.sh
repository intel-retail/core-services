#!/bin/bash
#
# Copyright (C) 2024 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

pipelineZooModel="https://github.com/dlstreamer/pipeline-zoo-models/raw/main/storage/"
modelPrecisionFP16INT8="FP16-INT8"

# $1 model file name
# $2 download URL
# $3 model percision
getModelFiles() {
    # Make model directory
    # ex. kdir efficientnet-b0/1/FP16-INT8
    mkdir -p "$1"/"$3"/1
    
    # Get the models
    wget "$2"/"$3"/"$1"".bin" -P "$1"/"$3"/1
    wget "$2"/"$3"/"$1"".xml" -P "$1"/"$3"/1
}

# $1 model file name
# $2 download URL
# $3 json file name
# $4 process file name (this can be different than the model name ex. horizontal-text-detection-0001 is using horizontal-text-detection-0002.json)
# $5 precision folder
getProcessFile() {
    # Get process file
    wget "$2"/"$3".json -O "$1"/"$5"/1/"$4".json
}

# custom yolov5s model downloading for this particular precision FP16INT8
downloadYolov5sFP16INT8() {
    yolov5s="yolov5s"
    yolojson="yolo-v5"

    # checking whether the model file .bin already exists or not before downloading
    yolov5ModelFile="models/$yolov5s/$modelPrecisionFP16INT8/1/$yolov5s.bin"
    if [ -f "$yolov5ModelFile" ]; then
        echo "yolov5s $modelPrecisionFP16INT8 model already exists in $yolov5ModelFile, skip downloading..."
    else
        echo "Downloading yolov5s $modelPrecisionFP16INT8 models..."
        # Yolov5s FP16_INT8
        getModelFiles $yolov5s $pipelineZooModel"yolov5s-416_INT8" $modelPrecisionFP16INT8
        getProcessFile $yolov5s $pipelineZooModel"yolov5s-416" $yolojson $yolov5s $modelPrecisionFP16INT8
        echo "yolov5 $modelPrecisionFP16INT8 model downloaded in $yolov5ModelFile"
    fi
}

downloadYolov5sFP16INT8