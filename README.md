# Cpu temp linker
I use i3status. It can't handle dynamically changing(on boot) paths for reading  
the cpu temp (ie. `/sys/class/hwmon/hwmon{N}/temp1_input`).  
This program *should* solve that issue...

## Instructions
Create a `config.json` file containg the sensor name and input label,  
that would correspond to `hwmon{N}/name` and `hwmon{N}/temp{N}_label`.  
For example:
```json
{
    "Cpu_sensor": "k10temp",
    "Cpu_input_label": "Tctl"
}
```

Then direct i3status to use the generated symlink. (currently `temp_input`)  
Now you should have a stable cpu temp probe :D

## Todo
- Better sensor path searching (ie. for drivers that don't use hwmon class)
- configurable symlink path
- ???
