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
echo "END GPU GET"
# ---------------------------------------

error() {
    printf '%s\n' "$1" >&2
    exit 1
}

show_help() {
	echo "
         usage: [RENDER_MODE=0 or 1]
                [DEVICE=\"CPU\" or other device] [MQTT=127.0.0.1:1883 if using MQTT broker] [COLOR_HEIGHT=1080] [COLOR_WIDTH=1920] [COLOR_FRAMERATE=15]
                sudo -E ./run.sh --profile \"object_detection\"(or others from make list-profiles) --target_device CPU|GPU|MULTI:GPU,CPU --inputsrc RS_SERIAL_NUMBER|CAMERA_RTSP_URL|file:video.mp4|/dev/video0

         Note: 
         1.  dgpu.x should be replaced with targeted GPUs such as dgpu (for all GPUs), dgpu.0, dgpu.1, etc
         2.  core.x should be replaced with targeted GPUs such as core (for all GPUs), core.0, core.1, etc
         3.  filesrc will utilize videos stored in the sample-media folder
         4.  when using device camera like USB, put your correspondent device number for your camera like /dev/video2 or /dev/video4
         5.  Set environment variable STREAM_DENSITY_MODE=1 for starting pipeline stream density testing
         6.  Set environment variable RENDER_MODE=1 for displaying pipeline and overlay CV metadata
         9.  Set environment variable PIPELINE_PROFILE=\"object_detection\" to run ovms pipeline profile object detection: values can be listed by \"make list-profiles\"
         10. Set environment variable STREAM_DENSITY_FPS=15.0 for setting stream density target fps value
         11. Set environment variable STREAM_DENSITY_INCREMENTS=1 for setting incrementing number of pipelines for running stream density
         13. Set environment variable MQTT=127.0.0.1:1883 for exporting inference metadata to an MQTT broker.
         14. Set environment variable like COLOR_HEIGHT, COLOR_WIDTH, and COLOR_FRAMERATE(in FPS) for inputsrc RealSense camera use cases.
        "
}

while :; do
    case $1 in
    -h | -\? | --help)
        show_help
        exit
        ;;
    --target_device)
        if [ "$2" ]; then
            if [ $2 == "CPU" ]; then
                TARGET_DEVICE=$2
                shift
            elif grep -q "core" <<< "$2"; then
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
            elif grep -q "GPU" <<< "$2"; then			
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
    --profile)
        if [ "$2" ]; then
            PROFILE=$2
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
# ---------------------------------------

echo "running pipeline profile: $PROFILE"

cl_cache_dir=$(pwd)/.cl-cache
echo "CLCACHE: $cl_cache_dir"

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

# Mount the USB if using a usb camera
TARGET_USB_DEVICE=""
if [[ "$INPUTSRC" == *"/dev/vid"* ]]
then
	TARGET_USB_DEVICE="--device=$INPUTSRC"
fi

# PROFILE is the environment variable to choose which type of pipelines to run with
# eg. grpc_python, grpc_cgo_binding, ... etc
# one example to run with this pipeline profile on the command line is like:
# PROFILE="grpc_python" sudo -E ./run.sh --platform core --inputsrc rtsp://127.0.0.1:8554/camera_0
echo "starting profile-launcher with pipeline profile: $PROFILE ..."
current_time=$(date "+%Y.%m.%d-%H.%M.%S")
cameras="${cameras[*]}" \
TARGET_USB_DEVICE="$TARGET_USB_DEVICE" \
TARGET_GPU_DEVICE="$TARGET_GPU_DEVICE" \
MQTT="$MQTT" \
RENDER_MODE=$RENDER_MODE \
DISPLAY=$DISPLAY \
cl_cache_dir=$cl_cache_dir \
RUN_PATH=$(pwd) \
BARCODE_RECLASSIFY_INTERVAL=$BARCODE_INTERVAL \
OCR_RECLASSIFY_INTERVAL=$OCR_INTERVAL \
OCR_DEVICE=$OCR_DEVICE \
LOG_LEVEL=$LOG_LEVEL \
INPUTSRC="$INPUTSRC" \
STREAM_DENSITY_FPS=$STREAM_DENSITY_FPS \
STREAM_DENSITY_INCREMENTS=$STREAM_DENSITY_INCREMENTS \
COMPLETE_INIT_DURATION=$COMPLETE_INIT_DURATION \
CPU_ONLY="$CPU_ONLY" \
PROFILE="$PROFILE" \
AUTO_SCALE_FLEX_140="$AUTO_SCALE_FLEX_140" \
./profile-launcher -configDir "$(dirname "$(readlink ./profile-launcher)")" > ./results/profile-launcher."$current_time".log 2>&1 &
