#Input: {
    Name:    string // The name of the MIDI input devices (available USB devices are printed on app startup)
    Channel: uint & < 16 // The MIDI channel to listen to (zero based, 0-15)
}

#Output: {
    Name:   ("xr" | "qu") // The TYPE name of mixing console to control (e.g. xr, qu)
    Ip:     string // The IP of mixing console to control (e.g. 192.168.0.150)
    Port:   int // The PORT of mixing console to control (e.g. 51325 or 10024)
}

#Mapping: {
    Name:    ("MuteGroup" | "TapDelay") // The FEATURE name to control (e.g. MuteGroup, TapDelay)
    Target:  uint // Index of the targeted feature (e.g. 1st MuteGroup, 3rd FX/TapDelay, 1-n)
    CC:      uint & < 128 // The MIDI CC to be triggered by (e.g. 0..127)
    ValueOn: uint & < 128 | *127 // The MIDI CC VALUE to be considered as "on" or "true" (e.g. relevant for MuteGroup)
}

Mappings: [...#Mapping]

