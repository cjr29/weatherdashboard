# Weather Dashboard
This program will connect to an MQTT broker and subscribe to multiple topics.
Each topic should be a message stream from one or more weather sensors.
The message is in JSON, which is parsed and made visible to the user in a data display.
The user can select which sensors to monitor and record.
The user can name the sensors and give them locations for easier identification.
# Technical
The program is configured by a **onfig.jsonc** file located in the local directory.
The GUI is configured first to support status reporting.
Once the GUI is set up, the program connects to the broker and subscribes to the
specified topics. Each topic is assigned a source 'home' to help the user identify
where a specific message originated.

The program is a single package, **main**, with four files, **main.go**, **datastructures.go**,
**dashboard.go**, **menuhandlers.go**, **theme**.go, **weatherwidget.go**, **config.go**,
and **messagehandling.go**. 

Configuration information is initialized using a **config.json** file,
and if configuration is changed, the configuration will be saved back into the **config.json** file.
Passwords and broker names are not saved in the code to ensure security. The config.json file will have the
passwords and broker names, so DO NOT upload config.json files to github!

A separate test file, **main_test.go** is provided to test the map functions.
