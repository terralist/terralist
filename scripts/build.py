import sys

if sys.version_info.major < 3 or sys.version_info.minor < 6:
  print ("ERROR: This build script requires Python 3.6+.")
  sys.exit(1)

from subprocess import check_output, run
from datetime import datetime
import os

SCRIPT_HOME = os.path.dirname(os.path.realpath(__file__))
PROJECT_HOME = os.path.abspath(os.path.join(SCRIPT_HOME, ".."))

BANNER="""
  _____                   _ _     _     ____        _ _     _           
 |_   _|__ _ __ _ __ __ _| (_)___| |_  | __ ) _   _(_) | __| | ___ _ __ 
   | |/ _ \ '__| '__/ _` | | / __| __| |  _ \| | | | | |/ _` |/ _ \ '__|
   | |  __/ |  | | | (_| | | \__ \ |_  | |_) | |_| | | | (_| |  __/ |   
   |_|\___|_|  |_|  \__,_|_|_|___/\__| |____/ \__,_|_|_|\__,_|\___|_|   
"""

if __name__ == "__main__":
  print (BANNER)
  
  if len(sys.argv) != 2 or sys.argv[1] not in ["debug", "release"]:
    print (f"USAGE: {sys.argv[0]} debug|release")
    sys.exit(1)

  mode = sys.argv[1]

  branch = check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"]).decode('utf-8').strip()
  version = f"{branch}-dev"
  if mode == "debug":
    version = f"{version}-debug"
  
  print (f"Version: {version}")

  commit_hash = check_output(["git", "rev-parse", "--short", "HEAD"]).decode('utf-8').strip()
  print (f"Commit Hash: {commit_hash}")

  build_timestamp = datetime.now().strftime("%Y-%m-%dT%H:%M:%S")
  print (f"Build Timestamp: {build_timestamp}")

  flags = " ".join([
    f"-X '{k}={v}'" 
    for k, v in {
      f"main.Version": version,
      f"main.CommitHash": commit_hash,
      f"main.BuildTimestamp": build_timestamp,
      f"main.Mode": mode,
    }.items()
  ])

  print ("")
  print ("Strating the build process...")
  print ("")

  binary_file_name = f"terralist{'.exe' if os.name == 'nt' else ''}"

  run(
    [
      "go", 
      "build", 
      f'-o={os.path.join(PROJECT_HOME, binary_file_name)}', 
      "-v", 
      f'-ldflags={flags}', 
      os.path.join(PROJECT_HOME, "cmd", "terralist", "main.go")
    ],
    stdin=sys.stdin,
    stdout=sys.stdout,
    stderr=sys.stderr
  )