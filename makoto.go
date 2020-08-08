/*** * * * readmify
language: golang*/
/*
# Document

# makoto version 1.0.0-rc2

This is a tool used to extract files from a .GPK file of the game "[スクールデイズ](https://overflow.fandom.com/wiki/School_Days)" or encode files into a .GPK file

此工具可用来提取GPK文件中的内容,也可以用来打包GPK文件

## Installation 安装方法

1. `go get github.com/hex3c/makoto`
2. `wget https://raw.githubusercontent.com/hex3c/makoto/master/makoto.go && go run makoto.go`
3. download a compiled binary file from the release page 从release页面下载编译好的二进制文件

## Usage 使用方法

```
REM extracts Event03.GPK of hq on Windows
REM 在Windows系统上解包hq的Event03.GPK
SET GPKFile=Event03
SET GPKUnpack=1
SET GPKKey=82EE1DB357E92CC22F547B104C9A7549
go run makoto.go
```

```
# extracts Se01.GPK of shiny on Linux
# 在Linux系统上解包shiny的Se01.GPK
GPKFile=Se01 GPKUnpack=1 GPKKey=F0D0BC0554AC68A9F17C8E3D640BF3AA go run makoto.go
```

```
REM encodes files in .Voice00/ into Voice00.GPK for cross on Windows
REM 在Windows系统上把Voice00文件夹打包成Voice00.GPK供cross使用
SET GPKFile=Voice00
SET GPKUnpack=0
SET GPKKey=567C1B90B6FE3FDBB60679EACC11A04F
makoto
```

```
# encodes files in .Script/ into Script.GPK for cross on Linux
# 在Linux系统上把Script文件夹打包成Script.GPK供cross使用
GPKFile=Script GPKUnpack=0 GPKKey=567C1B90B6FE3FDBB60679EACC11A04F makoto
```

## Known GPK keys 已知的GPK密钥

```
hq=82EE1DB357E92CC22F547B104C9A7549
shiny=F0D0BC0554AC68A9F17C8E3D640BF3AA
cross=567C1B90B6FE3FDBB60679EACC11A04F
```

you can submit an issue if you know the key of summer or sd(original)

如果你知道summer或者原版sd的密码可以发个issue

## Copying

```
Copyright (c) 2020 hex3c <hex3c@outlook.com>

This work is free. You can redistribute it and/or modify it under the
terms of the Do What The Fuck You Want To Public License, Version 2,
as published by Sam Hocevar. See below for more details.


       DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE 
                   Version 2, December 2004 

Copyright (C) 2004 Sam Hocevar <sam@hocevar.net> 

Everyone is permitted to copy and distribute verbatim or modified 
copies of this license document, and changing it is allowed as long 
as the name is changed. 

           DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE 
  TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION 

 0. You just DO WHAT THE FUCK YOU WANT TO.
```

## Changlog 更新记录

### v0.0.1 2020-08-05
can extract GPK / 可以解包GPK

### v1.0.0-rc1 2020-08-07
can repack GPK / 可以打包GPK

### v1.0.0-rc2 2020-08-08
solves panic in some special cases / 解决了在某些特殊情况下panic的问题

*/
/*
# Code

## header & global var
```
*/
package main

import "io"
import "os"
import "fmt"
import "path"
import "bytes"
import "errors"
import "encoding/hex"
import "compress/zlib"
import "path/filepath"

var printfcount uint
/*
```
## function: DoEncodeGPK
```
*/
func DoEncodeGPK(GPKFile string, GPKKey []uint8) (err error) {
/*
```
use FORIS.exe as header of GPK file
```
*/
	os.RemoveAll(GPKFile + ".GPK")
	err = os.Rename("FORIS.exe", GPKFile + ".GPK")
	if(err != nil) {
		return err
	}
	GPK, err := os.OpenFile(GPKFile + ".GPK", os.O_RDWR | os.O_APPEND, 0666)
	if(err != nil) {
		GPK.Close()
		return err
	}
	GPKFI, err := GPK.Stat()
	if(err != nil) {
		GPK.Close()
		return err
	}
/*
```
open each file in the dir

never use DFLT in encoding mode
```
*/
	ptr := uint(0)
	GPKptr := uint(GPKFI.Size())
	var DecompressedIndex []byte
	var filecount uint
	filecount = 0
	err = filepath.Walk(GPKFile, func(path string, info os.FileInfo, err error) error {
		if(err != nil) {
			return err
		}
		if(!info.IsDir()) {
			filecount++
			FileName := []rune(path[len(GPKFile) + 1:])
			for i := 0; i < len(FileName); i++ {
				if(FileName[i] == '\\') {
					FileName[i] = '/'
				}
			}
			InFile, _ := os.Open(GPKFile + "/" + string(FileName))
			fmt.Printf("%d:\t%d:\tOpened input file:\t%s\n", printfcount, filecount, GPKFile + "/" + string(FileName)); printfcount++;
			InFI, _ := InFile.Stat()
			FileSize := uint(InFI.Size())
			buffer := make([]byte, FileSize - 1)
			_, err := InFile.ReadAt(buffer, 1)
			if(err != nil) {
				GPK.Close()
				InFile.Close()
				return err
			}
			GPK, err := os.OpenFile(GPKFile + ".GPK", os.O_WRONLY | os.O_APPEND, 0666)
			GPK.Write(buffer)
			DataPosition := uint(GPKptr)
			fmt.Printf("%d:\t%d:\tData position is:\t0x%08x\n", printfcount, filecount, DataPosition); printfcount++;
			GPKptr += (FileSize - 1)
			DataLength := uint(FileSize)
			fmt.Printf("%d:\t%d:\tRaw data length is:\t0x%08x\n", printfcount, filecount, DataLength); printfcount++;
			fmt.Printf("%d:\t%d:\tDFLT:\talways false\n", printfcount, filecount); printfcount++;
			fmt.Printf("%d:\t%d:\tDecompressed data length:\talways 0x%08x\n", printfcount, filecount, 0); printfcount++;
			FileNameSize := uint(len(FileName))
			DecompressedIndex = append(DecompressedIndex, byte(FileNameSize % 256), byte(FileNameSize / 256))
			ptr += 2
			for i := uint(0); i < FileNameSize; i++ {
				DecompressedIndex = append(DecompressedIndex, byte(FileName[i] % 256), byte(FileName[i] / 256))
				ptr += 2
			}
			DecompressedIndex = append(DecompressedIndex, byte(0), byte(0), byte(0), byte(0), byte(0), byte(0))
			ptr += 6
			DecompressedIndex = append(DecompressedIndex, byte(DataPosition % 256), byte((DataPosition / 256) % 256), byte((DataPosition / 65536) % 256), byte((DataPosition / 16777216) % 256), byte(DataLength % 256), byte((DataLength / 256) % 256), byte((DataLength / 65536) % 256), byte((DataLength / 16777216) % 256))
			ptr += 8
			DecompressedIndex = append(DecompressedIndex, byte(20), byte(20), byte(20), byte(20), byte(0), byte(0), byte(0), byte(0), byte(1))
			ptr += 9
			buffer = make([]byte, 1)
			_, err = InFile.ReadAt(buffer, 0)
			if(err != nil) {
				GPK.Close()
				InFile.Close()
				return err
			}
			InFile.Close()
			DecompressedIndex = append(DecompressedIndex, buffer[0])
			ptr++
		}
		return nil
	})
	fmt.Printf("%d:\tSuccessfully encoded %d files\n", printfcount, filecount); printfcount++;
	if(err != nil) {
		GPK.Close()
		return err
	}
/*
```
compress the index and write to GPK file
```
*/
	DecompressedIndex = append(DecompressedIndex, byte(0), byte(0), byte(0), byte(0), byte(0))
	DecompressedLength := uint(len(DecompressedIndex))
	fmt.Printf("%d:\tDecompressed index size is:\t0x%08x\n", printfcount, DecompressedLength); printfcount++;
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(DecompressedIndex)
	w.Close()
	CompressedSize := uint(len(in.Bytes())) + uint(4)
	fmt.Printf("%d:\tCompressed index size is:\t0x%08x\n", printfcount, CompressedSize); printfcount++;
	buffer := make([]byte, CompressedSize)
	buffer[0] = byte(DecompressedLength % 256)
	buffer[1] = byte((DecompressedLength / 256) % 256)
	buffer[2] = byte((DecompressedLength / 65536) % 256)
	buffer[3] = byte((DecompressedLength / 16777216) % 256)
	for i := uint(4); i <= CompressedSize - 1; i++ {
		buffer[i] = in.Bytes()[i - 4]
	}
	for i := uint(0); i <= CompressedSize - 1; i++ {
		buffer[i] ^= GPKKey[i % 16]
	}
	GPK.Write(buffer)
	buffer = make([]byte, 32)
	buffer, err = hex.DecodeString("53544b46696c6530504944580000000053544b46696c65305041434b46494c45")
	if(err != nil) {
		GPK.Close()
		return err
	}
	buffer[12] = byte(CompressedSize % 256)
	buffer[13] = byte((CompressedSize / 256) % 256)
	buffer[14] = byte((CompressedSize / 65536) % 256)
	buffer[15] = byte((CompressedSize / 16777216) % 256)
	GPK.Write(buffer)
	GPK.Close()
	GPK, err = os.Open(GPKFile + ".GPK")
	if(err != nil) {
		GPK.Close()
		return err
	}
	GPKFI, err = GPK.Stat()
	if(err != nil) {
		GPK.Close()
		return err
	}
	GPKSize := GPKFI.Size()
	fmt.Printf("%d:\tSuccessfully done write GPK file, size of %s.GPK is:\t%d\n", printfcount, GPKFile, GPKSize); printfcount++;
	GPK.Close()
	return nil;
}
/*
```
## function: DoDecodeGPK
```
*/
func DoDecodeGPK(GPKFile string, GPKKey []uint8) (err error) {
/*
```
open the GPK file and verify

must end with "STKFile0PIDX" + 4byte + "STKFile0PACKFILE"

```
*/
	GPK, err := os.Open(GPKFile + ".GPK")
	if(err != nil) {
		GPK.Close()
		return err
	}
	GPKFI, err := GPK.Stat()
	if(err != nil) {
		GPK.Close()
		return err
	}
	GPKSize := GPKFI.Size()
	fmt.Printf("%d:\tThe size of the %s.GPK is:\t%d\n", printfcount, GPKFile, GPKSize); printfcount++;
	buffer := make([]byte, 32)
	_, err = GPK.ReadAt(buffer, GPKSize - 32)
	if(err != nil) {
		GPK.Close()
		return err
	}
	CompressedSize := uint(buffer[12]) + (uint(buffer[13]) * 256) + (uint(buffer[14]) * 65536) + (uint(buffer[15]) * 16777216)
	buffer[12] = 0
	buffer[13] = 0
	buffer[14] = 0
	buffer[15] = 0
	if(hex.EncodeToString(buffer) != "53544b46696c6530504944580000000053544b46696c65305041434b46494c45") {
		err := errors.New("Invalid GPK file: not found STKFile0PIDXuintSTKFile0PACKFILE at end")
		GPK.Close()
		return err
	}
	fmt.Printf("%d:\tCompressed index size is:\t0x%08x\n", printfcount, CompressedSize); printfcount++;
	buffer = make([]byte, CompressedSize)
	_, err = GPK.ReadAt(buffer, GPKSize - 32 - int64(CompressedSize))
	if(err != nil) {
		GPK.Close()
		return err
	}
/*
```
decode and decompress index
```
*/
	for i := uint(0); i <= CompressedSize - 1; i++ {
		buffer[i] ^= GPKKey[i % 16]
	}
	var out bytes.Buffer
	r, err := zlib.NewReader(bytes.NewReader(buffer[4:]))
	if(err != nil) {
		GPK.Close()
		return err
	}
	io.Copy(&out, r)
	DecompressedIndex := out.Bytes()
	DecompressedIndexSize := uint(buffer[0]) + (uint(buffer[1]) * 256) + (uint(buffer[2]) * 65536) + (uint(buffer[3]) * 16777216)
	fmt.Printf("%d:\tDecompressed index size is:\t0x%08x\n", printfcount, DecompressedIndexSize); printfcount++;
	if(uint(len(DecompressedIndex)) != DecompressedIndexSize) {
		GPK.Close()
		return errors.New("Incorrect decompressed index size")
	}
	ptr := uint(0)
	err = os.RemoveAll(GPKFile)
	if(err != nil) {
		GPK.Close()
		return err
	}
/*
```
process each file

note: index structre (byte):
- file name length = 2
- file = 2 * (file name length)
- 0x00 = 6
- data position = 4
- raw data length = 4
- DFLT = 4
- decompressed data length = 4
- header length = 1
- header = (header length)

```
*/
	var filecount uint
	filecount = 0
	for {
		filecount++
		FileNameSize := uint16(DecompressedIndex[ptr]) + (uint16(DecompressedIndex[ptr + 1]) * 256)
		if (FileNameSize == 0) {
			break
		}
		ptr += 2
		FileName := make([]rune, FileNameSize)
		for i := uint16(0); i < FileNameSize; i++ {
			FileName[i] = rune(DecompressedIndex[ptr + 2 * uint(i)]) + (rune(DecompressedIndex[ptr + 2 * uint(i) + 1]) * 256)
		}
		_, err := os.Stat(path.Dir(GPKFile + "/" + string(FileName)))
		if(err != nil) {
			os.MkdirAll(path.Dir(GPKFile + "/" + string(FileName)), os.ModePerm)
		}
		fmt.Printf("%d:\t%d:\tCreat output file:\t%s\n", printfcount, filecount, GPKFile + "/" + string(FileName)); printfcount++;
		OutFile, err := os.Create(GPKFile + "/" + string(FileName))
		if(err != nil) {
			GPK.Close()
			return err
		}
		ptr += 2 * uint(FileNameSize)
		ptr += 6
		DataPosition := uint(DecompressedIndex[ptr]) + (uint(DecompressedIndex[ptr + 1]) * 256) + (uint(DecompressedIndex[ptr + 2]) * 65536) + (uint(DecompressedIndex[ptr + 3]) * 16777216)
		fmt.Printf("%d:\t%d:\tData position is:\t0x%08x\n", printfcount, filecount, DataPosition); printfcount++;
		ptr += 4
		DataLength := uint(DecompressedIndex[ptr]) + (uint(DecompressedIndex[ptr + 1]) * 256) + (uint(DecompressedIndex[ptr + 2]) * 65536) + (uint(DecompressedIndex[ptr + 3]) * 16777216)
		fmt.Printf("%d:\t%d:\tRaw data length is:\t0x%08x\n", printfcount, filecount, DataLength); printfcount++;
		ptr += 4
		DFLT := (DecompressedIndex[ptr] == byte(0x44))
		fmt.Printf("%d:\t%d:\tDFLT:\t%t\n", printfcount, filecount, DFLT); printfcount++;
		ptr += 4
		DecompressedDataLength := uint(DecompressedIndex[ptr]) + (uint(DecompressedIndex[ptr + 1]) * 256) + (uint(DecompressedIndex[ptr + 2]) * 65536) + (uint(DecompressedIndex[ptr + 3]) * 16777216)
		fmt.Printf("%d:\t%d:\tDecompressed data length is:\t0x%08x\n", printfcount, filecount, DecompressedDataLength); printfcount++;
		ptr += 4
		RawData := make([]byte, DataLength)
		if(DecompressedIndex[ptr] != 0) {
			copy(RawData, DecompressedIndex[ptr + 1:uint(DecompressedIndex[ptr]) + 1 + ptr])
		}
		_, err = GPK.ReadAt(RawData[uint(DecompressedIndex[ptr]):], int64(DataPosition))
		if(err != nil) {
			GPK.Close()
			return err
		}
		var DecompressedData []byte
		if(DFLT) {
			var data bytes.Buffer
			CompressedBytes := bytes.NewReader(RawData)
			r, err := zlib.NewReader(CompressedBytes)
			if(err != nil) {
				GPK.Close()
				return err
			}
			io.Copy(&data, r)
			DecompressedData = data.Bytes()
			if(uint(len(DecompressedData)) != DecompressedDataLength) {
				GPK.Close()
				return errors.New("Incorrect decompressed data size")
			}
		} else {
			DecompressedData = RawData
		}
		OutFile.Write(DecompressedData)
		OutFile.Close()
		ptr += uint(DecompressedIndex[ptr])
		ptr++
	}
	fmt.Printf("%d:\tSuccessfully decoded %d files\n", printfcount, filecount - 1); printfcount++;
/*
```
extract FORIS.exe for repacking
```
*/
	buffer = make([]byte, 5120)
	GPK.ReadAt(buffer, 0)
	os.RemoveAll("FORIS.exe")
	FORIS, _ := os.Create("FORIS.exe")
	FORIS.Write(buffer)
	GPK.Close()
	FORIS.Close()
	fmt.Printf("%d:\tExtracted FORIS.exe\n", printfcount); printfcount++;
	return nil
}
/*
```
## function PrintInfo()
```
*/
func PrintInfo() {
	InfoStr, _ := hex.DecodeString("6D616B6F746F2076657273696F6E20312E302E302D726332202868747470733A2F2F6769746875622E636F6D2F68657833632F6D616B6F746F290A0A436F70797269676874202863292032303230206865783363203C6865783363406F75746C6F6F6B2E636F6D3E0A0A5468697320776F726B20697320667265652E20596F752063616E2072656469737472696275746520697420616E642F6F72206D6F6469667920697420756E646572207468650A7465726D73206F662074686520446F205768617420546865204675636B20596F752057616E7420546F205075626C6963204C6963656E73652C2056657273696F6E20322C0A6173207075626C69736865642062792053616D20486F63657661722E205365652062656C6F7720666F72206D6F72652064657461696C732E0A0A0A20202020202020444F205748415420544845204655434B20594F552057414E5420544F205055424C4943204C4943454E5345200A2020202020202020202020202020202020202056657273696F6E20322C20446563656D6265722032303034200A0A436F707972696768742028432920323030342053616D20486F6365766172203C73616D40686F63657661722E6E65743E200A0A45766572796F6E65206973207065726D697474656420746F20636F707920616E64206469737472696275746520766572626174696D206F72206D6F646966696564200A636F70696573206F662074686973206C6963656E736520646F63756D656E742C20616E64206368616E67696E6720697420697320616C6C6F776564206173206C6F6E67200A617320746865206E616D65206973206368616E6765642E200A0A2020202020202020202020444F205748415420544845204655434B20594F552057414E5420544F205055424C4943204C4943454E5345200A20205445524D5320414E4420434F4E444954494F4E5320464F5220434F5059494E472C20444953545249425554494F4E20414E44204D4F44494649434154494F4E200A0A20302E20596F75206A75737420444F205748415420544845204655434B20594F552057414E5420544F2E0A0A0A7573616765206578616D706C65733A0A0A52454D2057696E646F77730A52454D206578747261637473204576656E7430332E47504B206F662068710A5345542047504B46696C653D4576656E7430330A5345542047504B556E7061636B3D310A5345542047504B4B65793D38324545314442333537453932434332324635343742313034433941373534390A676F2072756E206D616B6F746F2E676F0A52454D20656E636F6465732066696C657320696E202E566F69636530302F20696E746F20566F69636530302E47504B20666F722063726F73730A5345542047504B46696C653D566F69636530300A5345542047504B556E7061636B3D300A5345542047504B4B65793D35363743314239304236464533464442423630363739454143433131413034460A6D616B6F746F0A0A23204C696E75780A2320657874726163747320536530312E47504B206F66207368696E79206F6E204C696E75780A47504B46696C653D536530312047504B556E7061636B3D312047504B4B65793D463044304243303535344143363841394631374338453344363430424633414120676F2072756E206D616B6F746F2E676F0A2320656E636F6465732066696C657320696E202E5363726970742F20696E746F205363726970742E47504B20666F722063726F73730A47504B46696C653D5363726970742047504B556E7061636B3D302047504B4B65793D3536374331423930423646453346444242363036373945414343313141303446206D616B6F746F0A0A4B6E6F776E2047504B206B6579730A0A68713D38324545314442333537453932434332324635343742313034433941373534390A7368696E793D46304430424330353534414336384139463137433845334436343042463341410A63726F73733D35363743314239304236464533464442423630363739454143433131413034460A0A796F752063616E207375626D697420616E20697373756520696620796F75206B6E6F7720746865206B6579206F662073756D6D6572206F72207364286F726967696E616C290A0A0A")
	fmt.Println(string(InfoStr))
}
/*
```
## main function

decodes GPKKey from environment variables
```
*/
func main() {
	printfcount = 1
	GPKFile := os.Getenv("GPKFile")
	GPKUnpack := (os.Getenv("GPKUnpack") == "1")
	GPKKey, err := hex.DecodeString(os.Getenv("GPKKey"))
	if(err != nil) {
		PrintInfo()
		panic(err)
	}
	fmt.Printf("%d:\tDecoded GPKKey is:\t%v\n", printfcount, GPKKey); printfcount++;
	if(len(GPKKey) != 16) {
		PrintInfo()
		panic("GPKKey incorrect length")
	}
	if(GPKUnpack) {
		err := DoDecodeGPK(GPKFile, GPKKey)
		if(err != nil) {
			PrintInfo()
			panic(err)
		}
	} else {
		err := DoEncodeGPK(GPKFile, GPKKey)
		if(err != nil) {
			PrintInfo()
			panic(err)
		}
	}
}
/*
```
*/
