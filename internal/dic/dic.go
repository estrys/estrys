//nolint:ireturn
package dic

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

var ErrServiceAlreadyRegistred = fmt.Errorf("service already registered")

var services = make(map[string]any) //nolint:gochecknoglobals

func typeName[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().String()
}

func GetService[T any]() T {
	serviceName := typeName[T]()
	service, exist := services[serviceName]
	if !exist {
		panic(errors.Errorf("service %s does not exist", serviceName))
	}
	return service.(T) //nolint:forcetypeassert
}

func Register[T any](implementation T) error {
	if _, exist := services[typeName[T]()]; exist {
		return nil
	}
	services[typeName[T]()] = implementation
	return nil
}

func ResetContainer() {
	services = make(map[string]interface{}, len(services))
}
