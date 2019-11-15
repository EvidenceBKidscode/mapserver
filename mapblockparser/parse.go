package mapblockparser

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"mapserver/coords"
	"strconv"
)

func Parse(data []byte, mtime int64, pos *coords.MapBlockCoords) (*MapBlock, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	mapblock := NewMapblock()
	mapblock.Mtime = mtime
	mapblock.Pos = pos
	mapblock.RawData = data
	mapblock.Size = len(data)

	err := mapblock.Parse()
	if err != nil {
		return nil, err
	} else {
		return mapblock, nil
	}
}

func (mapblock *MapBlock)Parse() (error) {
	if mapblock.Parsed {
		return nil
	}

	mapblock.Size = len(mapblock.RawData)

	if mapblock.Size == 0 {
		return errors.New("no data")
	}

	timer := prometheus.NewTimer(parseDuration)
	defer timer.ObserveDuration()


	data := mapblock.RawData

	// version
	mapblock.Version = data[0]

	if mapblock.Version < 25 || mapblock.Version > 28 {
		return errors.New("mapblock-version not supported: " + strconv.Itoa(int(mapblock.Version)))
	}

	//flags
	flags := data[1]
	mapblock.Underground = (flags & 0x01) == 0x01

	var offset int

	if mapblock.Version >= 27 {
		offset = 4
	} else {
		//u16 lighting_complete not present
		offset = 2
	}

	content_width := data[offset]
	params_width := data[offset+1]

	if content_width != 2 {
		return errors.New("content_width = " + strconv.Itoa(int(content_width)))
	}

	if params_width != 2 {
		return errors.New("params_width = " + strconv.Itoa(int(params_width)))
	}

	//mapdata (blocks)
	if mapblock.Version >= 27 {
		offset = 6

	} else {
		offset = 4

	}

	//metadata
	count, err := parseMapdata(mapblock, data[offset:])
	if err != nil {
		return err
	}

	offset += count

	count, err = parseMetadata(mapblock, data[offset:])
	if err != nil {
		return err
	}

	offset += count

	//static objects

	offset++ //static objects version
	staticObjectsCount := readU16(data, offset)
	offset += 2
	for i := 0; i < staticObjectsCount; i++ {
		offset += 13
		dataSize := readU16(data, offset)
		offset += dataSize + 2
	}

	//timestamp
	offset += 4

	//mapping version
	offset++

	numMappings := readU16(data, offset)
	offset += 2
	for i := 0; i < numMappings; i++ {
		nodeId := readU16(data, offset)
		offset += 2

		nameLen := readU16(data, offset)
		offset += 2

		blockName := string(data[offset : offset+nameLen])
		offset += nameLen

		mapblock.BlockMapping[nodeId] = blockName
	}

	mapblock.Parsed = true
	mapblock.RawData = nil
	parsedMapBlocks.Inc()

	return nil
}
