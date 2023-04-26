#!/bin/bash

# Define the program name
startScript="start.sh"
programName="fadeout"
programDirectory="/usr/local/bin/Fadeout"

# Build the program
go build -o $programName main.go

# Generate start script
startScript="#!/bin/bash\n\
$programDirectory/fadeout"

echo -e $startScript > start.sh
chmod +x start.sh

# Copy the program files to the installation directory
sudo mkdir -p $programDirectory

sudo cp $programName "$programDirectory/"
sudo cp -r icons "$programDirectory/"
sudo cp -r configs "$programDirectory/"
sudo mv start.sh "$programDirectory/"

# Generate .desktop file
desktopFile="[Desktop Entry]\n\
Type=Application\n\
Name=$programName\n\
Icon=$programDirectory/icons/icon_100x100.png\n\
Exec=$programDirectory/start.sh\n\
Terminal=false\n\
Categories=Utility;"

echo -e $desktopFile > "$HOME/.local/share/applications/$programName.desktop"

sudo update-desktop-database

echo "Installation completed!"
