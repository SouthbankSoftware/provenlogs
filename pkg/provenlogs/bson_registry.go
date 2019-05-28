/*
 * @Author: guiguan
 * @Date:   2019-05-26T22:04:59+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-26T23:25:08+10:00
 */

package provenlogs

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

var (
	bsonRegistry     *bsoncodec.Registry
	bsonRegistryOnce = new(sync.Once)
)

// DefaultBSONRegistry returns the default BSON registry being used by ProvenLogs
func DefaultBSONRegistry() *bsoncodec.Registry {
	bsonRegistryOnce.Do(func() {
		bsonRegistry = NewOrderedBSONRegistry()
	})

	return bsonRegistry
}

// NewOrderedBSONRegistry creates a default BSON registry that encodes a map into an ordered BSON
// doc with its keys sorted in string order
func NewOrderedBSONRegistry() *bsoncodec.Registry {
	mapEncodeValue := func(ec bsoncodec.EncodeContext, dw bsonrw.DocumentWriter, val reflect.Value, collisionFn func(string) bool) error {
		encoder, err := ec.LookupEncoder(val.Type().Elem())
		if err != nil {
			return err
		}

		keys := val.MapKeys()
		sort.Slice(keys, func(i int, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		for _, key := range keys {
			if collisionFn != nil && collisionFn(key.String()) {
				return fmt.Errorf("Key %s of inlined map conflicts with a struct field name", key)
			}
			vw, err := dw.WriteDocumentElement(key.String())
			if err != nil {
				return err
			}

			if enc, ok := encoder.(bsoncodec.ValueEncoder); ok {
				err = enc.EncodeValue(ec, vw, val.MapIndex(key))
				if err != nil {
					return err
				}
				continue
			}
			err = encoder.EncodeValue(ec, vw, val.MapIndex(key))
			if err != nil {
				return err
			}
		}

		return dw.WriteDocumentEnd()
	}

	return bson.NewRegistryBuilder().RegisterDefaultEncoder(reflect.Map, bsoncodec.ValueEncoderFunc(func(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
		if !val.IsValid() || val.Kind() != reflect.Map || val.Type().Key().Kind() != reflect.String {
			return bsoncodec.ValueEncoderError{Name: "MapEncodeValue", Kinds: []reflect.Kind{reflect.Map}, Received: val}
		}

		if val.IsNil() {
			// If we have a nill map but we can't WriteNull, that means we're probably trying to encode
			// to a TopLevel document. We can't currently tell if this is what actually happened, but if
			// there's a deeper underlying problem, the error will also be returned from WriteDocument,
			// so just continue. The operations on a map reflection value are valid, so we can call
			// MapKeys within mapEncodeValue without a problem.
			err := vw.WriteNull()
			if err == nil {
				return nil
			}
		}

		dw, err := vw.WriteDocument()
		if err != nil {
			return err
		}

		return mapEncodeValue(ec, dw, val, nil)
	})).Build()
}
