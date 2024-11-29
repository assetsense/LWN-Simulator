import json

# Load JSON data from a file
input_file = 'components/devices-2000.json'   # Replace with your file path
output_file = 'components/devices.json'

with open(input_file, 'r') as file:
    data = json.load(file)

# filtered_data = {key: value for key, value in data.items() if value['info'].get('axisId') != 6169}

# for key,value in data.items():

file_path = 'device_euis.txt'  # You can change the file path as needed
with open(file_path, 'w') as f:
    for key, value in data.items():
        f.write(f"{value['info'].get('devEUI')}\n")


# Save the filtered data to a new file
# with open(output_file, 'w') as file:
#     json.dump(filtered_data, file, indent=4)

print(f"Filtered data has been saved")