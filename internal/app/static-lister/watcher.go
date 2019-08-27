/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package static_lister

// Watcher fills the gauge vector and watches for changes in the label file

import (
	"bufio"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type Watcher struct {
	labelFile string
	gauge *prometheus.GaugeVec

	currentLabels []string
}

func NewWatcher(labelFile string, gauge *prometheus.GaugeVec) (*Watcher, error) {
	w := &Watcher{
		labelFile: labelFile,
		gauge: gauge,
		currentLabels: []string{},
	}

	err := w.updateLabels()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) Run(errChan chan<- error) {
	log.Debug().Msg("starting watcher")

	notifier, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error().Err(err).Msg("error initializing fsnotify")
		errChan <- err
		return
	}
	defer notifier.Close()

	err = notifier.Add(w.labelFile)
	if err != nil {
		log.Error().Err(err).Str("file", w.labelFile).Msg("error watching file")
		errChan <- err
		return
	}

	for {
		select {
		case event, ok := <-notifier.Events:
			if !ok {
				log.Warn().Msg("notifier event channel closed; stopping watcher")
				errChan <- nil
				return
			}

			log.Debug().Interface("event", event).Msg("received event")
			if event.Op & fsnotify.Rename == fsnotify.Rename || event.Op & fsnotify.Remove == fsnotify.Remove {
				// Rename or delete, likely as part of saving from an editor.
				// Set up new watcher.
				err := notifier.Add(w.labelFile)
				if err != nil {
					log.Error().Err(err).Str("file", w.labelFile).Msg("error watching file")
					errChan <- err
					return
				}
				continue
			}

			// Any other event we see as a trigger to reload
			err := w.updateLabels()
			if err != nil {
				log.Error().Err(err).Msg("error updating labels")
				errChan <- err
				return
			}
		case err, ok := <-notifier.Errors:
			if !ok {
				log.Warn().Msg("notifier error channel closed; stopping watcher")
				errChan <- nil
				return
			}
			log.Warn().Err(err).Msg("error watching; continuing anyway")
		}
	}
}

func (w *Watcher) updateLabels() error {
	log.Debug().Msg("updating labels")

	// Try to read label file
	file, err := os.Open(w.labelFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var newLabels []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		newLabels = append(newLabels, scanner.Text())
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	// Delete old labels
	log.Debug().Strs("labels", w.currentLabels).Msg("deleting current labels")
	for _, label := range(w.currentLabels) {
		w.gauge.DeleteLabelValues(label)
	}

	// Create new labels
	log.Debug().Strs("labels", newLabels).Msg("adding new labels")
	for _, label := range(newLabels) {
		g, err := w.gauge.GetMetricWithLabelValues(label)
		if err != nil {
			return err
		}
		g.Set(1) // We just set the metric to 1, as in "has to exist == true"
	}

	w.currentLabels = newLabels
	return nil
}
