import os
import re
import json
import sys

# Allowed labels for the curated release notes
ALLOWED_LABELS = {
    "kind/bug-fix",
    "kind/feature",
    "kind/upgrade-consideration",
    "kind/breaking-change",
    "kind/api-change",
    "kind/deprecation",
    "impact/high",
    "impact/medium"
}

# Regex pattern for folder names (vX.X.X)
VERSION_PATTERN = re.compile(r'^v\d+\.\d+\.\d+$')

def validate_folder_name(path):
    folder_name = os.path.basename(path)
    if not VERSION_PATTERN.match(folder_name):
        print(f"Error: Folder '{folder_name}' does not match the required version format 'vX.X.X'.")
        return False
    return True

def validate_labels(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
        for note in data.get('notes', []):
            for label in note.get('label', []):
                if label not in ALLOWED_LABELS:
                    print(f"Error: Invalid label '{label}' in file '{file_path}'.")
                    return False
    return True

def lint_files(files):
    valid = True
    for file in files:
        if 'raw' in file or 'curated' in file:
            folder_name = os.path.basename(os.path.dirname(file))
            if not validate_folder_name(folder_name):
                valid = False

            if 'curated' in file and file.endswith('.json'):
                if not validate_labels(file):
                    valid = False

    return valid

if __name__ == "__main__":
    changed_files = sys.argv[1:]
    if not lint_files(changed_files):
        sys.exit(1)
    else:
        print("All checks passed successfully.")
