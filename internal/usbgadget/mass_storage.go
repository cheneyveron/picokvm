package usbgadget

var massStorageBaseConfig = gadgetConfigItem{
	order:      3000,
	device:     "mass_storage.usb0",
	path:       []string{"functions", "mass_storage.usb0"},
	configPath: []string{"mass_storage.usb0"},
	attrs: gadgetAttributes{
		"stall": "1",
	},
}

var massStorageLun0Config = gadgetConfigItem{
	order: 3001,
	path:  []string{"functions", "mass_storage.usb0", "lun.0"},
	attrs: gadgetAttributes{
		"cdrom":     "1",
		"ro":        "1",
		"removable": "1",
		"file":      "\n",
		// the additional whitespace is intentional to avoid the "KVM V irtual Media" string
		// Vendor (8 chars), product (16 chars)
		"inquiry_string": "KVM  Virtual Media",
	},
}

var mtpConfig = gadgetConfigItem{
	order:      3003,
	device:     "ffs.mtp",
	path:       []string{"functions", "ffs.mtp"},
	configPath: []string{"ffs.mtp"},
}

// SetDefaultMassStorageCdrom sets the cdrom mode the gadget is first bound
// with. Call it before NewUsbGadget: if the first image mounted at startup
// needs the other mode, changing it later rebinds the UDC, and the host sees
// the KVM unplugged and replugged a few seconds after it first appeared, which
// a BIOS that is already running can fail to pick up (keyboard dead).
func SetDefaultMassStorageCdrom(cdrom bool) {
	mode := "0"
	if cdrom {
		mode = "1"
	}
	massStorageLun0Config.attrs["cdrom"] = mode
	if item, ok := defaultGadgetConfig["mass_storage_lun0"]; ok {
		item.attrs["cdrom"] = mode
	}
}
