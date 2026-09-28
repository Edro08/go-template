package config

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// cloneValue realiza una copia profunda de los valores internos. Previene que mapas y slices retornados mantengan referencias a c.data.
func cloneValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		cp := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			cp[k] = cloneValue(v2)
		}
		return cp
	case []interface{}:
		cp := make([]interface{}, len(val))
		for i, v2 := range val {
			cp[i] = cloneValue(v2)
		}
		return cp
	default:
		// Los tipos primitivos (string, int, bool, etc.) en Go se pasan por valor.
		return val
	}
}

// normalizeValue normaliza valores externos a la representación interna utilizada por Config.
// Los mapas se convierten a map[string]interface{} y los slices/arrays a []interface{} de forma recursiva.
func normalizeValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Map:
		result := make(map[string]interface{}, rv.Len())

		for _, key := range rv.MapKeys() {
			if key.Kind() != reflect.String {
				return nil, fmt.Errorf("%w: got %s", ErrMapKeyType, key.Kind())
			}

			value, err := normalizeValue(
				rv.MapIndex(key).Interface(),
			)
			if err != nil {
				return nil, err
			}

			result[key.String()] = value
		}

		return result, nil

	case reflect.Slice, reflect.Array:
		result := make([]interface{}, rv.Len())

		for i := 0; i < rv.Len(); i++ {
			value, err := normalizeValue(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}

			result[i] = value
		}

		return result, nil

	default:
		return v, nil
	}
}

// findValue navega el mapa interno. Debe ser llamado únicamente bajo un bloqueo (RLock o Lock).
func (c *Config) findValue(key string) (interface{}, bool) {
	if key == "" {
		return nil, false
	}

	keys := strings.Split(key, c.opts.Separator)
	var current interface{} = c.data

	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[k]
		if !exists {
			return nil, false
		}
		current = val
	}

	return current, true
}

// getValue adquiere el RLock de forma segura, busca el valor y lo clona antes de liberarlo.
func (c *Config) getValue(key string) (interface{}, bool) {
	val, ok := c.findValue(key)
	if !ok {
		return nil, false
	}

	return cloneValue(val), true
}

// getMap devuelve una copia segura del mapa interno.
func (c *Config) getMap(key string) (map[string]interface{}, bool) {
	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	m, ok := v.(map[string]interface{})
	return m, ok
}

// getSlice devuelve una copia segura del slice interno.
func (c *Config) getSlice(key string) ([]interface{}, bool) {
	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	s, ok := v.([]interface{})
	return s, ok
}

func (c *Config) set(key string, value interface{}) error {
	if key == "" {
		return ErrKeyEmpty
	}

	value, err := normalizeValue(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	keys := strings.Split(key, c.opts.Separator)
	cm := c.data

	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]

		next, exists := cm[k]

		if exists {
			if m, ok := next.(map[string]interface{}); ok {
				cm = m
				continue
			}
		}

		nm := make(map[string]interface{})
		cm[k] = nm
		cm = nm
	}

	cm[keys[len(keys)-1]] = value

	return nil
}

// convertToStringMap convierte un map[string]interface{} a map[string]string,
// aplicando la conversión a string para cada valor.
func convertToStringMap(input map[string]interface{}) (map[string]string, bool) {
	data := make(map[string]string)
	for k, v := range input {
		data[k] = toString(v)
	}
	return data, true
}

// convertToIntMap convierte un map[string]interface{} a map[string]int,
// aplicando la conversión a int para cada valor.
func convertToIntMap(input map[string]interface{}) (map[string]int, bool) {
	data := make(map[string]int)
	for k, v := range input {
		data[k] = toInt(v)
	}
	return data, true
}

// convertToFloatMap convierte un map[string]interface{} a map[string]float64,
// aplicando la conversión a float64 para cada valor.
func convertToFloatMap(input map[string]interface{}) (map[string]float64, bool) {
	data := make(map[string]float64)
	for k, v := range input {
		data[k] = toFloat64(v)
	}
	return data, true
}

// convertToBoolMap convierte un map[string]interface{} a map[string]bool,
// aplicando la conversión a bool para cada valor.
func convertToBoolMap(input map[string]interface{}) (map[string]bool, bool) {
	data := make(map[string]bool)
	for k, v := range input {
		data[k] = toBool(v)
	}
	return data, true
}

// convertToStringSlice convierte un slice []interface{} a []string,
// aplicando la conversión a string para cada elemento.
func convertToStringSlice(input []interface{}) ([]string, bool) {
	result := make([]string, 0, len(input))
	for _, v := range input {
		result = append(result, toString(v))
	}
	return result, true
}

// convertToIntSlice convierte un slice []interface{} a []int,
// aplicando la conversión a int para cada elemento.
func convertToIntSlice(input []interface{}) ([]int, bool) {
	result := make([]int, 0, len(input))
	for _, v := range input {
		result = append(result, toInt(v))
	}
	return result, true
}

// convertToFloatSlice convierte un slice []interface{} a []float64,
// aplicando la conversión a float64 para cada elemento.
func convertToFloatSlice(input []interface{}) ([]float64, bool) {
	result := make([]float64, 0, len(input))
	for _, v := range input {
		result = append(result, toFloat64(v))
	}
	return result, true
}

// convertToBoolSlice convierte un slice []interface{} a []bool,
// aplicando la conversión a bool para cada elemento.
func convertToBoolSlice(input []interface{}) ([]bool, bool) {
	result := make([]bool, 0, len(input))
	for _, v := range input {
		result = append(result, toBool(v))
	}
	return result, true
}

// toString convierte cualquier valor básico a string,
// manejando tipos numéricos, booleanos y cadenas.
// Para tipos no reconocidos devuelve cadena vacía.
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

// toInt convierte un valor básico a int, realizando conversiones
// seguras desde tipos numéricos y cadenas. Retorna 0 si no es convertible.
func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		if strconv.IntSize == 32 && (v < math.MinInt32 || v > math.MaxInt32) {
			return 0
		}
		return int(v)
	case uint:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0
		}
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0
		}
		return int(v)
	case uint64:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0
		}
		if strconv.IntSize == 64 && v > math.MaxInt64 {
			return 0
		}
		return int(v)
	case float64:
		if v != math.Trunc(v) {
			return 0
		}
		if strconv.IntSize == 32 {
			if v < math.MinInt32 || v > math.MaxInt32 {
				return 0
			}
		} else {
			if v < math.MinInt64 || v > math.MaxInt64 {
				return 0
			}
		}
		return int(v)

	case float32:
		f := float64(v)
		if f != math.Trunc(f) {
			return 0
		}
		if strconv.IntSize == 32 {
			if f < math.MinInt32 || f > math.MaxInt32 {
				return 0
			}
		} else {
			if f < math.MinInt64 || f > math.MaxInt64 {
				return 0
			}
		}
		return int(v)

	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return i

	default:
		return 0
	}
}

// toFloat64 convierte un valor básico a float64, incluyendo conversiones
// desde enteros, flotantes y cadenas.
// Retorna 0 si no es convertible.
func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

// toBool convierte un valor básico a bool, aceptando valores booleanos
// y cadenas "true"/"false" (sin distinguir mayúsculas).
// Retorna false si no es convertible.
func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		if strings.EqualFold(v, "true") {
			return true
		} else if strings.EqualFold(v, "false") {
			return false
		}
		return false
	default:
		return false
	}
}
