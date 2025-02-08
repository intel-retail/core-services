# 
# Copyright (C) 2025 Intel Corporation. 
# 
# SPDX-License-Identifier: Apache-2.0 
#

import pyudev
import subprocess
import json
import os
import cv2
import glob
import platform
from flask import jsonify
import ast  # Added import for ast.literal_eval
os.environ["OPENCV_AVFOUNDATION_SKIP_AUTH"] = "1"
# Dummy camera data
dummy_cameras = {
    "camera_001": {
        "id": "camera_001",
        "type": "wired",
        "connection": "USB",
        "index": 0,  # Add this field
        "status": "active",
        "name": "Logitech HD Webcam",
        "resolution": "1920x1080",
        "fps": 30,
        "ip": None,  # Wired cameras don't have an IP
        "port": None  # Wired cameras don't have a port
    },
    "camera_002": {
        "id": "camera_002",
        "type": "wireless",
        "connection": "Wi-Fi",
        "index": 1,  # Add this field
        "status": "active",
        "name": "Arlo Pro 3",
        "resolution": "2560x1440",
        "fps": 25,
        "ip": "192.168.1.102",
        "port": 554
    },
    "camera_003": {
        "id": "camera_003",
        "type": "wired",
        "connection": "HDMI",
        "index": 2,  # Add this field
        "status": "inactive",
        "name": "Sony Alpha a6400",
        "resolution": "3840x2160",
        "fps": 60,
        "ip": None,
        "port": None
    }
}

def get_dummy_cameras():
    """
    Returns dummy data for connected cameras as a dictionary.

    """
    return dummy_cameras


def get_available_cameras(start_index=0):
    """
    Scans the system for available cameras and returns a list of dictionaries.
    Works across Linux, macOS, and Windows.
    """
    cameras = []
    system = platform.system()

    if system == "Linux":
        # Use /dev/video* to detect available cameras
        video_devices = sorted(glob.glob("/dev/video*"))
        for device in video_devices:
            index = int(device.replace("/dev/video", ""))
            camera_name, resolution, fps = get_camera_properties(index)

            if camera_name:
                cameras.append({
                    "id": f"camera_00{start_index}",
                    "type": "wired",
                    "connection": "integrated",  # Assuming USB as the default
                    "index": index,
                    "status": "active",
                    "name": camera_name,
                    "resolution": resolution,
                    "fps": fps,
                    "ip": None,  # Wired cameras don't have an IP
                    "port": None,  # Wired cameras don't have a port
                })
                start_index += 1

    else:  # macOS & Windows (no /dev/video*)
        for index in range(10):  # Check first 10 indices
            camera_name, resolution, fps = get_camera_properties(index)
            if camera_name:
                cameras.append({
                    "id": f"camera_00{start_index}",
                    "type": "wired",
                    "connection": "integrated",
                    "index": index,
                    "status": "active",
                    "name": camera_name,
                    "resolution": resolution,
                    "fps": fps,
                    "ip": None,
                    "port": None,
                })
                start_index += 1

    return cameras


def scan_wired_cameras(start_index=1):
    """
    Scans the system for wired cameras using OpenCV and retrieves basic camera information.
    Args:
        start_index (int): The starting index for camera numbering.
    Returns:
        list: A list of wired camera information.
        int: The next available index after processing wired cameras.
    """
    cameras = []

    # Iterate over a range of possible camera indices
    for camera_index in range(10):  # Check the first 10 indices (adjust as needed)
        cap = cv2.VideoCapture(camera_index)
        if cap.isOpened():
            # Fetch basic details about the camera
            camera_name = f"Camera {camera_index}"
            resolution = f"{int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))}x{int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))}"
            fps = int(cap.get(cv2.CAP_PROP_FPS))

            cameras.append({
                "id": f"camera_00{start_index}",
                "type": "wired",
                "connection": "integrated",  # Assuming USB as the default for OpenCV
                "index": start_index,
                "status": "active",
                "name": camera_name,
                "resolution": resolution,
                "fps": fps,
                "ip": None,  # Wired cameras don't have IP
                "port": None,  # Wired cameras don't have a port
            })
            start_index += 1  # Increment the shared index
            cap.release()

    return cameras, start_index

def scan_local_cameras(start_index):
    """
    Scans the system for locally connected cameras (e.g., USB, built-in webcams).
    Args:
        start_index (int): The starting index for camera numbering.
    Returns:
        list: A list of local camera information.
        int: The next available index after processing local cameras.
    """
    cameras = []
    video_devices = sorted(glob.glob("/dev/video*"))  # Get available video devices

    for device in video_devices:
        print(f"Checking device: {device}")
        index = int(device.replace("/dev/video", ""))
        cap = cv2.VideoCapture(index)

        if cap.isOpened():
            cameras.append({
                "id": f"camera_00{start_index}",
                "type": "USB",
                "connection": "Wired",
                "index": index,
                "status": "active",
                "name": f"Local Camera {index}",
                "resolution": get_camera_resolution(cap),
                "fps": get_camera_fps(cap),
                "device": device
            })
            cap.release()
        else:
            print(f"Warning: Cannot access {device}. It may be in use.")

        start_index += 1

    return cameras, start_index

def scan_network_cameras(start_index):
    """
    Scans the network for wireless cameras using Nmap.
    Args:
        start_index (int): The starting index for camera numbering.
    Returns:
        list: A list of network camera information.
        int: The next available index after processing network cameras.
    """
    cameras = []
    try:
        # Use a list of arguments instead of shell=True for security
        nmap_args = [
            "nmap",
            "-T5",
            "-p", "554,80",
            "-oG", "-",
            "192.168.0.0/24"  # Consider making this configurable
        ]
        result = subprocess.run(
            nmap_args,
            capture_output=True,
            text=True,
            check=True  # Raise CalledProcessError if return code is non-zero
        )
        output = result.stdout

        # Parse Nmap output for active cameras
        for line in output.split("\n"):
            if "open" in line:
                parts = line.split()
                ip = parts[1]
                cameras.append({
                    "id": f"camera_00{start_index}",
                    "type": "wireless",
                    "connection": "Wi-Fi",
                    "index": start_index,  # Removed duplicate index
                    "status": "active",
                    "name": "Unknown Network Camera",
                    "resolution": "Unknown",
                    "fps": None,
                    "ip": ip,
                    "port": 554 if "554/open" in line else 80,
                })
                start_index += 1
    except subprocess.CalledProcessError as e:
        print(f"Error running nmap: {e}")
    except Exception as e:
        print(f"Error scanning network: {e}")

    return cameras, start_index

def store_cameras_to_file(cameras, file_name="scanned_cameras.txt"):
    """
    Stores the camera data in a text file in the specified format.
    
    Args:
        cameras (list): List of camera dictionaries.
        file_name (str): Name of the file to store the data. Default is 'scanned_cameras.txt'.
    """
    # Convert the list of cameras into a dictionary with keys like "camera_001"
    cameras_dict = {f"camera_{str(i+1).zfill(3)}": cam for i, cam in enumerate(cameras)}
    
    # Write the cameras to the file in the desired format
    with open(file_name, "w") as file:
        file.write("actual_cameras = {\n")
        for key, camera in cameras_dict.items():
            file.write(f'    "{key}": {json.dumps(camera, indent=4)},\n')
        file.write("}\n")

def read_actual_cameras(file_path):
    """
    Reads a file containing a Python dictionary (with assignment) and parses it into a Python object.

    Args:
        file_path (str): Path to the .txt file.

    Returns:
        dict: A dictionary representation of the cameras data.
    """
    abs_path = os.path.abspath(file_path)
    
    if not os.path.exists(abs_path):
        print(f"Error: {abs_path} not found!")
        return []
        
    with open(file_path, "r") as file:
        content = file.read()

    # Replace `null` with `None` to make it compatible with Python
    content = content.replace("null", "None")

    # Extract the dictionary part after `actual_cameras =`
    content = content.split("=", 1)[1].strip()

    # Safely evaluate the content into a Python dictionary using ast.literal_eval
    try:
        actual_cameras = ast.literal_eval(content)
    except (ValueError, SyntaxError) as e:
        print(f"Error parsing camera data: {e}")
        return {}

    return actual_cameras

def get_camera_resolution(cap):
    """Get camera resolution."""
    width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    return f"{width}x{height}" if width and height else "Unknown"

def get_camera_fps(cap):
    """Get camera FPS (frames per second)."""
    fps = cap.get(cv2.CAP_PROP_FPS)
    return fps if fps > 0 else "Unknown"

def test_camera(index):
    """Tests if the camera at the given index is accessible."""
    cap = cv2.VideoCapture(index)
    if cap.isOpened():
        cap.release()
        return True
    return False

def get_camera_properties(index):
    """Attempts to get camera properties like name, resolution, and FPS."""
    cap = cv2.VideoCapture(index)

    if not cap.isOpened():
        return None, None, None

    resolution = f"{int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))}x{int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))}"
    fps = cap.get(cv2.CAP_PROP_FPS)
    camera_name = f"Local Camera {index}"

    cap.release()
    return camera_name, resolution, fps if fps > 0 else "Unknown"