#!/bin/bash
#
# Copyright (C) 2023 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

# shellcheck source=/dev/null
HAS_FLEX_140=0
HAS_FLEX_170=0
HAS_ARC=0
GPU_NUM_140=0
GPU_NUM_170=0

has_flex_170=`lspci -d :56C0`
has_flex_140=`lspci -d :56C1`
has_arc=`lspci | grep -iE "5690|5691|5692|56A0|56A1|56A2|5693|5694|5695|5698|56A5|56A6|56B0|56B1|5696|5697|56A3|56A4|56B2|56B3"`

if [ -z "$has_flex_170" ] && [ -z "$has_flex_140" ] && [ -z "$has_arc" ] 
then
	echo "No discrete Intel GPUs found"
	return
fi
echo "GPU exists!"

if [ ! -z "$has_flex_140" ]
then
	HAS_FLEX_140=1
	GPU_NUM_140=`echo "$has_flex_140" | wc -l`
fi
if [ ! -z "$has_flex_170" ]
then
	HAS_FLEX_170=1
	GPU_NUM_170=`echo "$has_flex_170" | wc -l`
fi
if [ ! -z "$has_arc" ]
then
	HAS_ARC=1
fi

echo "HAS_FLEX_140=$HAS_FLEX_140, HAS_FLEX_170=$HAS_FLEX_170, HAS_ARC=$HAS_ARC, GPU_NUM_140=$GPU_NUM_140, GPU_NUM_170=$GPU_NUM_170"


if [ -z "$PLATFORM" ] || [ -z "$INPUTSRC" ]
then
# shellcheck source=/dev/null
    while :; do
        case $1 in
        -h | -\? | --help)
            show_help
            exit
            ;;
        --platform)
            if [ "$2" ]; then
                if [ $2 == "xeon" ]; then
                    PLATFORM=$2
                    shift
                elif grep -q "core" <<< "$2"; then
                    PLATFORM="core"
                    arrgpu=(${2//./ })
                    TARGET_GPU_NUMBER=${arrgpu[1]}
                    if [ -z "$TARGET_GPU_NUMBER" ]; then
                        TARGET_GPU="GPU.0"
                        # TARGET_GPU_DEVICE="--privileged"
                    else
                        TARGET_GPU_ID=$((128+$TARGET_GPU_NUMBER))
                        TARGET_GPU="GPU."$TARGET_GPU_NUMBER
                        TARGET_GPU_DEVICE="--device=/dev/dri/renderD"$TARGET_GPU_ID
                        TARGET_GPU_DEVICE_NAME="/dev/dri/renderD"$TARGET_GPU_ID
                    fi
                    echo "CORE"
                    shift
                elif grep -q "dgpu" <<< "$2"; then			
                    PLATFORM="dgpu"
                    arrgpu=(${2//./ })
                    TARGET_GPU_NUMBER=${arrgpu[1]}
                    if [ -z "$TARGET_GPU_NUMBER" ]; then
                        TARGET_GPU="GPU.0"
                        # TARGET_GPU_DEVICE="--privileged"
                    else
                        TARGET_GPU_ID=$((128+$TARGET_GPU_NUMBER))
                        TARGET_GPU="GPU."$TARGET_GPU_NUMBER
                        TARGET_GPU_DEVICE="--device=/dev/dri/renderD"$TARGET_GPU_ID
                        TARGET_GPU_DEVICE_NAME="/dev/dri/renderD"$TARGET_GPU_ID
                    fi
                    #echo "$PLATFORM $TARGET_GPU"
                    shift	
                else
                    error 'ERROR: "--platform" requires an argument core|xeon|dgpu.'
                fi
            else
                    error 'ERROR: "--platform" requires an argument core|xeon|dgpu.'
            fi	    
            ;;
        --inputsrc)
            if [ "$2" ]; then
                INPUTSRC=$2
                shift
            else
                error 'ERROR: "--inputsrc" requires an argument RS_SERIAL_NUMBER|CAMERA_RTSP_URL|file:video.mp4|/dev/video0.'
            fi
            ;;
        -?*)
            error "ERROR: Unknown option $1"
            ;;
        ?*)
            error "ERROR: Unknown option $1"
            ;;
        *)
            break
            ;;
        esac

        shift

    done

    if [ -z $PLATFORM ] || [ -z $INPUTSRC ]
    then
        #echo "Blanks: $1 $PLATFORM $INPUTSRC"
        show_help
        exit 0
    fi

    echo "Device:"
    echo $TARGET_GPU_DEVICE
fi

echo "running pipeline profile: $PIPELINE_PROFILE"

SERVER_CONTAINER_NAME="ovms-server"
# clean up exited containers
exitedOvmsContainers=$(docker ps -a -f name="$SERVER_CONTAINER_NAME" -f status=exited -q)
if [ -n "$exitedOvmsContainers" ]
then
	docker rm "$exitedOvmsContainers"
fi

export GST_DEBUG=0

cl_cache_dir=$(pwd)/.cl-cache
echo "CLCACHE: $cl_cache_dir"


if [ "$HAS_FLEX_140" == 1 ] || [ "$HAS_FLEX_170" == 1 ] || [ "$HAS_ARC" == 1 ] 
then
    echo "OCR device defaulting to dGPU"
    OCR_DEVICE=GPU
fi

gstDockerImages="dlstreamer:dev"
re='^[0-9]+$'
if grep -q "video" <<< "$INPUTSRC"; then
	echo "assume video device..."
	# v4l2src /dev/video*
	# TODO need to pass stream info
	TARGET_USB_DEVICE="--device=$INPUTSRC"
elif [[ "$INPUTSRC" =~ $re ]]; then
	echo "assume realsense device..."
	cameras=$(find /dev/vid* | while read -r line; do echo "--device=$line"; done)
	# replace \n with white space as the above output contains \n
	cameras=("$(echo "$cameras" | tr '\n' ' ')")
	TARGET_GPU_DEVICE=$TARGET_GPU_DEVICE" ""${cameras[*]}"
	gstDockerImages="dlstreamer:realsense"
else
	echo "$INPUTSRC"
fi

if [ "$PIPELINE_PROFILE" == "gst" ]
then
	# modify gst profile DockerImage accordingly based on the inputsrc is RealSense camera or not

	docker run --rm -v "${PWD}":/workdir mikefarah/yq -i e '.OvmsClient.DockerLauncher.DockerImage |= "'"$gstDockerImages"'"' \
		/workdir/configs/opencv-ovms/cmd_client/res/gst/configuration.yaml
fi

#pipeline script is configured from configuration.yaml in opencv-ovms/cmd_client/res folder

# Set RENDER_MODE=1 for demo purposes only
if [ "$RENDER_MODE" == 1 ]
then
	xhost +local:docker
fi

#echo "DEBUG: $TARGET_GPU_DEVICE $PLATFORM $HAS_FLEX_140"
if [ "$PLATFORM" == "dgpu" ] && [ "$HAS_FLEX_140" == 1 ]
then
	if [ "$STREAM_DENSITY_MODE" == 1 ]; then
		AUTO_SCALE_FLEX_140=2
	else
		# allow workload to manage autoscaling
		AUTO_SCALE_FLEX_140=1
	fi
fi

# make sure models are downloaded or existing:
./download_models/getModels.sh

# make sure sample image is downloaded or existing:
./configs/opencv-ovms/scripts/image_download.sh

# devices supported CPU, GPU, GPU.x, AUTO, MULTI:GPU,CPU
DEVICE="${DEVICE:="CPU"}"

# PIPELINE_PROFILE is the environment variable to choose which type of pipelines to run with
# eg. grpc_python, grpc_cgo_binding, ... etc
# one example to run with this pipeline profile on the command line is like:
# PIPELINE_PROFILE="grpc_python" sudo -E ./run.sh --platform core --inputsrc rtsp://127.0.0.1:8554/camera_0
PIPELINE_PROFILE="${PIPELINE_PROFILE:=grpc_python}"
echo "starting profile-launcher with pipeline profile: $PIPELINE_PROFILE ..."
current_time=$(date "+%Y.%m.%d-%H.%M.%S")
cameras="${cameras[*]}" \
TARGET_USB_DEVICE="$TARGET_USB_DEVICE" \
TARGET_GPU_DEVICE="$TARGET_GPU_DEVICE" \
DEVICE="$DEVICE" \
MQTT="$MQTT" \
RENDER_MODE=$RENDER_MODE \
DISPLAY=$DISPLAY \
cl_cache_dir=$cl_cache_dir \
RUN_PATH=$(pwd) \
PLATFORM=$PLATFORM \
BARCODE_RECLASSIFY_INTERVAL=$BARCODE_INTERVAL \
OCR_RECLASSIFY_INTERVAL=$OCR_INTERVAL \
OCR_DEVICE=$OCR_DEVICE \
LOG_LEVEL=$LOG_LEVEL \
LOW_POWER="$LOW_POWER" \
STREAM_DENSITY_MODE=$STREAM_DENSITY_MODE \
INPUTSRC="$INPUTSRC" \
STREAM_DENSITY_FPS=$STREAM_DENSITY_FPS \
STREAM_DENSITY_INCREMENTS=$STREAM_DENSITY_INCREMENTS \
COMPLETE_INIT_DURATION=$COMPLETE_INIT_DURATION \
CPU_ONLY="$CPU_ONLY" \
PIPELINE_PROFILE="$PIPELINE_PROFILE" \
AUTO_SCALE_FLEX_140="$AUTO_SCALE_FLEX_140" \
./profile-launcher -configDir "$(dirname "$(readlink ./profile-launcher)")" > ./results/profile-launcher."$current_time".log 2>&1 &