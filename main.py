import os
import requests

def get_command(query):
  """Gets a kubectl command from OpenAI GPT."""

  url = "https://api.openai.com/v1/engines/gpt/completions"
  headers = {"Authorization": "Bearer YOUR_API_KEY"}
  data = {"prompt": query}

  response = requests.post(url, headers=headers, data=data)
  if response.status_code == 200:
    return response.json()["choices"][0]["text"]
  else:
    raise Exception("Error getting command from OpenAI GPT: {}".format(response.status_code))

def main():
  """Main function."""

  query = input("What do you want to do? ")
  command = get_command(query)
  print(command)

if __name__ == "__main__":
  main()

import os
import subprocess

def execute_command(command):
  """Executes a kubectl command."""

  process = subprocess.Popen(
      command,
      stdout=subprocess.PIPE,
      stderr=subprocess.PIPE,
      shell=True)
  stdout, stderr = process.communicate()
  if process.returncode != 0:
    raise Exception("Error executing kubectl command: {}".format(stderr))
  else:
    return stdout

def main():
  """Main function."""

  command = input("What kubectl command do you want to execute? ")
  output = execute_command(command)
  print(output)

if __name__ == "__main__":
  main()

# Sources:
# 1. https://github.com/floyd-fuh/JKS-private-key-cracker-hashcat
# 2. https://github.com/OsamaJBR/Beringei-HTTP-Interface