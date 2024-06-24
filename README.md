# Weather Dashboard
This program will connect to an MQTT broker and subscribe to multiple topics.
Each topic should be a message stream from one or more weather sensors.
The message is in JSON, which is parsed and made visible to the user in a data display.
The user can select which sensors to monitor and record.
The user can name the sensors and give them locations for easier identification.
# Technical
The program is configured by a **config.ini** file located in the local directory.
The GUI is configured first to support status reporting.
Once the GUI is set up, the program connects to the broker and subscribes to the
specified topics. Each topic is assigned a source 'home' to help the user identify
where a specific message originated.

The program is a single package, **main**, with four files, **main.go**, **datastructures.go**,
**config.go**, and **messagehandling**. Configuration information is initialized using a .ini file,
but when user-configurable menu is developed, the configuration will be saved in a .yaml file.
Passwords and broker names are not saved in the code to ensure security. The .yaml file will have the
passwords and broker names, so DO NOT upload .yaml files to github!

A separate test file, **main_test.go** is provided to test the map functions.
