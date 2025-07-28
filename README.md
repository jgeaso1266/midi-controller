# Module midi-controller 

This module provides MIDI input functionality for Viam robots, allowing them to receive and process MIDI messages from external MIDI controllers, keyboards, and other MIDI devices. The module acts as a sensor that continuously listens to MIDI input and tracks which keys are currently being pressed, making it useful for music-based robotics applications, interactive installations, or any scenario where MIDI input should control robot behavior.

## Model jalen:midi-controller:midi-input-reader

This model implements a MIDI input reader that connects to a specified MIDI input port and continuously monitors for MIDI note events. It tracks which keys are currently pressed (note on events) and removes them when released (note off events). The sensor provides real-time access to the currently active MIDI keys through the Viam sensor API.

The implementation uses the [gomidi](https://gitlab.com/gomidi/midi) library to handle MIDI communication and runs a background goroutine to listen for MIDI messages. All MIDI data access is thread-safe using mutex locks.

### Configuration
The following attribute template can be used to configure this model:

```json
{
  "in_port_name": "<string>"
}
```

#### Attributes

The following attributes are available for this model:

| Name          | Type   | Inclusion | Description                |
|---------------|--------|-----------|----------------------------|
| `in_port_name` | string  | Required  | The name of the MIDI input port to connect to. This should match exactly with the MIDI port name as recognized by your system (e.g., "USB MIDI Device", "Piano", etc.) |

#### Example Configuration

```json
{
  "in_port_name": "USB MIDI Device"
}
```

### Sensor Readings

The sensor returns the following data when `Readings()` is called:

- `keys`: A comma-separated string of currently pressed MIDI key numbers (0-127). Returns an empty string when no keys are pressed.

### Usage Notes

- The MIDI port must be available and accessible when the component starts
- Only note on/off events are currently tracked - other MIDI messages are ignored  
- The component automatically handles port opening/closing and cleanup
- Key numbers follow standard MIDI convention (0-127, where middle C is typically 60)

### Example Robot Configuration

```json
{
  "components": [
    {
      "name": "midi_input",
      "model": "jalen:midi-controller:midi-input-reader",
      "type": "sensor",
      "attributes": {
        "in_port_name": "USB MIDI Device"
      }
    }
  ]
}
```