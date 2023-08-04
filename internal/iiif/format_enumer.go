// Code generated by "enumer -type format"; DO NOT EDIT.

package iiif

import (
	"fmt"
	"strings"
)

const _formatName = "jpgtifpnggifjp2pdfwebp"

var _formatIndex = [...]uint8{0, 3, 6, 9, 12, 15, 18, 22}

const _formatLowerName = "jpgtifpnggifjp2pdfwebp"

func (i format) String() string {
	if i >= format(len(_formatIndex)-1) {
		return fmt.Sprintf("format(%d)", i)
	}
	return _formatName[_formatIndex[i]:_formatIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _formatNoOp() {
	var x [1]struct{}
	_ = x[jpg-(0)]
	_ = x[tif-(1)]
	_ = x[png-(2)]
	_ = x[gif-(3)]
	_ = x[jp2-(4)]
	_ = x[pdf-(5)]
	_ = x[webp-(6)]
}

var _formatValues = []format{jpg, tif, png, gif, jp2, pdf, webp}

var _formatNameToValueMap = map[string]format{
	_formatName[0:3]:        jpg,
	_formatLowerName[0:3]:   jpg,
	_formatName[3:6]:        tif,
	_formatLowerName[3:6]:   tif,
	_formatName[6:9]:        png,
	_formatLowerName[6:9]:   png,
	_formatName[9:12]:       gif,
	_formatLowerName[9:12]:  gif,
	_formatName[12:15]:      jp2,
	_formatLowerName[12:15]: jp2,
	_formatName[15:18]:      pdf,
	_formatLowerName[15:18]: pdf,
	_formatName[18:22]:      webp,
	_formatLowerName[18:22]: webp,
}

var _formatNames = []string{
	_formatName[0:3],
	_formatName[3:6],
	_formatName[6:9],
	_formatName[9:12],
	_formatName[12:15],
	_formatName[15:18],
	_formatName[18:22],
}

// formatString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func formatString(s string) (format, error) {
	if val, ok := _formatNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _formatNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to format values", s)
}

// formatValues returns all values of the enum
func formatValues() []format {
	return _formatValues
}

// formatStrings returns a slice of all String values of the enum
func formatStrings() []string {
	strs := make([]string, len(_formatNames))
	copy(strs, _formatNames)
	return strs
}

// IsAformat returns "true" if the value is listed in the enum definition. "false" otherwise
func (i format) IsAformat() bool {
	for _, v := range _formatValues {
		if i == v {
			return true
		}
	}
	return false
}
