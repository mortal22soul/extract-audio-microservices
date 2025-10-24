#!/usr/bin/env python3
"""
Minimal test to debug the quality analyzer import issue
"""
import sys
import os
import traceback

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

print("Testing minimal imports...")

try:
    print("1. Importing basic modules...")
    import logging
    import asyncio
    import numpy as np
    import cv2
    print("✓ Basic modules imported")
    
    print("2. Importing config...")
    from analytics.config import config
    print("✓ Config imported")
    
    print("3. Testing class definition...")
    
    # Define a minimal version of the class
    class TestVideoQualityAnalyzer:
        def __init__(self):
            self.test = True
    
    test_analyzer = TestVideoQualityAnalyzer()
    print("✓ Test class created successfully")
    
    print("4. Testing the actual file execution...")
    
    # Read and execute the file line by line to find the issue
    with open('src/analytics/quality_analyzer.py', 'r') as f:
        content = f.read()
    
    print(f"File size: {len(content)} characters")
    
    # Try to execute the content
    try:
        exec(content)
        print("✓ File executed successfully")
        
        # Check if classes are defined
        if 'VideoQualityAnalyzer' in locals():
            print("✓ VideoQualityAnalyzer class found")
        else:
            print("❌ VideoQualityAnalyzer class not found")
            
        if 'ContentModerationAnalyzer' in locals():
            print("✓ ContentModerationAnalyzer class found")
        else:
            print("❌ ContentModerationAnalyzer class not found")
            
    except Exception as e:
        print(f"❌ Error executing file: {e}")
        traceback.print_exc()
        
        # Try to find the problematic line
        lines = content.split('\n')
        for i, line in enumerate(lines[:50]):  # Check first 50 lines
            try:
                exec(line)
            except Exception as line_error:
                print(f"Error on line {i+1}: {line}")
                print(f"Error: {line_error}")
                break
    
except Exception as e:
    print(f"❌ Error: {e}")
    traceback.print_exc()