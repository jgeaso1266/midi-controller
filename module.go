package midicontroller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils/rpc"
)

var (
	MidiInputReader = resource.NewModel("jalen", "midi-controller", "midi-input-reader")
	errNotFound     = errors.New("not found")
)

func init() {
	resource.RegisterComponent(sensor.API, MidiInputReader,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newMidiControllerMidiInputReader,
		},
	)
}

type Config struct {
	InPortName string `json:"in_port_name"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	// Add config validation code here
	if cfg.InPortName == "" {
		return nil, nil, errNotFound
	}

	return nil, nil, nil
}

type midiMessage struct {
	NotesInfo map[uint8]uint8
}

type midiControllerMidiInputReader struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	inPort       drivers.In
	midiReadings midiMessage
	mu           sync.RWMutex
}

func (s *midiControllerMidiInputReader) listenToMidiInput() {
	s.logger.Infof("Starting MIDI input listener for port: %s", s.cfg.InPortName)

	if err := s.inPort.Open(); err != nil {
		s.logger.Errorf("Failed to open MIDI input port %s: %v", s.cfg.InPortName, err)
		return
	}
	defer s.inPort.Close() // Ensure the port is closed when the goroutine exits

	stopFn, err := midi.ListenTo(s.inPort, func(msg midi.Message, timestampms int32) {
		var ch, key, vel uint8
		s.mu.Lock()         // Lock the mutex before writing to readings
		defer s.mu.Unlock() // Unlock after writing

		switch {
		case msg.GetNoteStart(&ch, &key, &vel):
			s.logger.Infof("MIDI Note Start: %d", key)
			s.midiReadings.NotesInfo[key] = vel
		case msg.GetNoteEnd(&ch, &key):
			s.logger.Infof("MIDI Note End: %d", key)
			delete(s.midiReadings.NotesInfo, key)
		}
	}, midi.UseSysEx())

	if err != nil {
		s.logger.Errorf("Failed to start MIDI listener for port %s: %v", s.cfg.InPortName, err)
		return
	}
	defer stopFn() // Stop the MIDI listening when the goroutine exits

	s.logger.Info("MIDI input listener stopped.")
}

func newMidiControllerMidiInputReader(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewMidiInputReader(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewMidiInputReader(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (sensor.Sensor, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &midiControllerMidiInputReader{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		midiReadings: midiMessage{
			NotesInfo: make(map[uint8]uint8),
		},
	}

	var err error
	s.inPort, err = midi.FindInPort(conf.InPortName)
	if err != nil {
		return nil, fmt.Errorf("failed to find MIDI input port '%s': %w", conf.InPortName, err)
	}

	// Start the MIDI listener in a Goroutine
	go s.listenToMidiInput()

	return s, nil
}

func (s *midiControllerMidiInputReader) Name() resource.Name {
	return s.name
}

func (s *midiControllerMidiInputReader) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (sensor.Sensor, error) {
	client := &midiControllerMidiInputReader{
		name:   name,
		logger: logger,
		cfg:    s.cfg,
	}

	// You can add additional logic here if needed to initialize the client
	return client, nil
}

func (s *midiControllerMidiInputReader) toString(m map[uint8]uint8) (string, string) {
	if len(m) == 0 {
		return "", ""
	}

	keys := make([]string, len(m))
	velocities := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = fmt.Sprintf("%d", k)
		velocities[i] = fmt.Sprintf("%d", m[k])
		i++
	}
	return strings.Join(keys, ", "), strings.Join(velocities, ", ")
}

func (s *midiControllerMidiInputReader) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock() // Read lock for concurrent safe access
	defer s.mu.Unlock()
	// Return a copy of the readings to prevent external modification
	copiedReadings := make(map[string]interface{})

	copiedReadings["keys"], copiedReadings["velocities"] = s.toString(s.midiReadings.NotesInfo)
	return copiedReadings, nil
}

func (s *midiControllerMidiInputReader) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}

func (s *midiControllerMidiInputReader) Close(context.Context) error {
	midi.CloseDriver()
	s.cancelFunc()
	return nil
}
