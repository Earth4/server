/*
	Copyright 2012-2013 1620469 Ontario Limited.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"encoding/xml"
	"fmt"
	"os"
)

import (
	"gk/gkerr"
)

type GameConfigDef struct {
	XMLName                xml.Name `xml:"config"`
	HttpPort               int      `xml:"httpPort"`
	TokenPort              int      `xml:"tokenPort"`
	WebsocketPort          int      `xml:"websocketPort"`
	LogDir                 string   `xml:"logDir"`
	TemplateDir            string   `xml:"templateDir"`
	AvatarSvgDir           string   `xml:"avatarSvgDir"`
	TerrainSvgDir          string   `xml:"terrainSvgDir"`
	WebAddressPrefix       string   `xml:"webAddressPrefix"`
	WebsocketAddressPrefix string   `xml:"websocketAddressPrefix"`
	AudioAddressPrefix     string   `xml:"audioAddressPrefix"`
	DatabaseHost           string   `xml:"databaseHost"`
	DatabasePort           int      `xml:"databasePort"`
	DatabaseUserName       string   `xml:"databaseUserName"`
	DatabasePassword       string   `xml:"databasePassword"`
	DatabaseDatabase       string   `xml:"databaseDatabase"`
	WebsocketPath          string   `xml:"websocketPath"`
	CertificatePath        string   `xml:"certificatePath"`
	PrivateKeyPath         string   `xml:"privateKeyPath"`
}

func LoadConfigFile(fileName string) (*GameConfigDef, *gkerr.GkErrDef) {
	var err error
	var gameConfig *GameConfigDef = new(GameConfigDef)

	var file *os.File
	file, err = os.Open(fileName)
	if err != nil {
		return gameConfig, gkerr.GenGkErr(fmt.Sprintf("os.Open file: %s", fileName), err, ERROR_ID_OPEN_CONFIG)
	}
	defer file.Close()

	err = xml.NewDecoder(file).Decode(gameConfig)
	if err != nil {
		return gameConfig, gkerr.GenGkErr(fmt.Sprintf("xml.NewDecoder file: %s", fileName), err, ERROR_ID_DECODE_CONFIG)
	}

	return gameConfig, nil
}
