package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"unsafe"
)

type (
	Info struct {
		Subdir string `json:"subdir"`
	}

	Package struct {
		Build       string   `json:"build"`
		BuildNumber uint32   `json:"build_number"`
		Depends     []string `json:"depends"`
		License     string   `json:"license"`
		MD5         string   `json:"md5"`
		Name        string   `json:"name"`
		// Noarch      bool   `json: "noarch"`
		Sha256    string `json:"sha256"`
		Size      uint32 `json:"size"`
		Subdir    string `json:"subdir"`
		Timestamp uint64 `json:"timestamp"`
		Version   string `json:"version"`
	}

	PackageConda struct {
		Package

		Constrains    []string `json:"constrains"`
		LegacyBz2Md5  string   `json:"legacy_bz2_md5"`
		LicenseFamily string   `json:"license_family"`
	}
)

type CondaRepoDataJSON struct {
	Info            Info                    `json:"info"`
	Packages        map[string]Package      `json:"packages"`
	PackagesConda   map[string]PackageConda `json:"packages.conda"`
	Removed         []string                `json:"removed"`
	RepoDataVersion uint32                  `json:"repodata_version"`
}

// Convert maps to arrays since it isn't supported yet
type CondaRepoData struct {
	Info            Info
	Packages        []Package
	PackagesConda   []PackageConda
	Removed         []string
	RepoDataVersion uint32
}

var (
	repoData      CondaRepoData
	repoDataBytes []byte
	repoDataJSON  bytes.Buffer
)

func init() {
	resp, err := http.Get("https://conda.anaconda.org/conda-forge/noarch/current_repodata.json")
	if err != nil {
		panic(fmt.Sprintf("failed bench data download: %s", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("failed bench data download with status code %d", resp.StatusCode))
	}
	t := io.TeeReader(resp.Body, &repoDataJSON)
	d := json.NewDecoder(t)
	var rd CondaRepoDataJSON
	err = d.Decode(&rd)
	if err != nil {
		panic(fmt.Sprintf("failed json decoding: %s", err))
	}
	repoData.Info.Subdir = rd.Info.Subdir
	repoData.Removed = rd.Removed
	repoData.RepoDataVersion = rd.RepoDataVersion
	repoData.Packages = make([]Package, 0, len(rd.Packages))
	for _, p := range rd.Packages {
		repoData.Packages = append(repoData.Packages, p)
	}
	repoData.PackagesConda = make([]PackageConda, 0, len(rd.PackagesConda))
	for _, p := range rd.PackagesConda {
		repoData.PackagesConda = append(repoData.PackagesConda, p)
	}

	var buf bytes.Buffer
	repoData.WriteTo(&buf)
	repoDataBytes = buf.Bytes()
}

func BenchmarkCondaBinencRead(b *testing.B) {
	b.SetBytes(int64(len(repoDataBytes)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var rd CondaRepoData
		rd.ReadFrom(bytes.NewReader(repoDataBytes))
	}
}

func BenchmarkCondaBinencWrite(b *testing.B) {
	var buf bytes.Buffer
	b.ResetTimer()
	size := 0
	for i := 0; i < b.N; i++ {
		n, _ := repoData.WriteTo(&buf)
		size = n
		b.SetBytes(int64(n))
		buf.Reset()
	}
	b.Logf("binenc write size: %d\n", size)
}

func BenchmarkCondaJSONRead(b *testing.B) {
	b.ResetTimer()
	b.SetBytes(int64(repoDataJSON.Len()))
	for i := 0; i < b.N; i++ {
		var rd CondaRepoData
		e := json.NewDecoder(bytes.NewReader(repoDataJSON.Bytes()))
		e.Decode(&rd)
		b.SetBytes(int64(repoDataJSON.Len()))
	}
}

func BenchmarkCondaJSONWrite(b *testing.B) {
	var buf bytes.Buffer
	b.ResetTimer()
	size := 0
	for i := 0; i < b.N; i++ {
		e := json.NewEncoder(&buf)
		e.Encode(repoData)
		size = buf.Len()
		b.SetBytes(int64(size))
		buf.Reset()
	}
	b.Logf("json write size: %d\n", size)
}

func (s *CondaRepoData) WriteTo(w io.Writer) (n int, err error) {
	size := 12
	size += len(s.Info.Subdir)
	for _, v := range s.Removed {
		size += 2
		size += len(v)
	}
	for _, v := range s.PackagesConda {
		size += 38
		size += len(v.Package.Build) + len(v.Package.License) + len(v.Package.MD5) + len(v.Package.Name) + len(v.Package.Sha256) + len(v.Package.Subdir) + len(v.Package.Version) + len(v.LegacyBz2Md5) + len(v.LicenseFamily)
		for _, v1 := range v.Constrains {
			size += 2
			size += len(v1)
		}
		for _, v1 := range v.Package.Depends {
			size += 2
			size += len(v1)
		}
	}
	for _, v := range s.Packages {
		size += 32
		size += len(v.Build) + len(v.License) + len(v.MD5) + len(v.Name) + len(v.Sha256) + len(v.Subdir) + len(v.Version)
		for _, v1 := range v.Depends {
			size += 2
			size += len(v1)
		}
	}
	buf := make([]byte, size)
	offset := 0
	buf[offset] = byte(len(s.Info.Subdir))
	buf[offset+1] = byte(len(s.Info.Subdir) >> 8)
	offset += 2
	copy(buf[offset:], s.Info.Subdir)
	offset += len(s.Info.Subdir)
	buf[offset] = byte(len(s.Packages))
	buf[offset+1] = byte(len(s.Packages) >> 8)
	offset += 2
	for _, v := range s.Packages {
		buf[offset] = byte(len(v.Build))
		buf[offset+1] = byte(len(v.Build) >> 8)
		offset += 2
		copy(buf[offset:], v.Build)
		offset += len(v.Build)
		buf[offset] = byte(v.BuildNumber)
		buf[offset+1] = byte(v.BuildNumber >> 8)
		buf[offset+2] = byte(v.BuildNumber >> 16)
		buf[offset+3] = byte(v.BuildNumber >> 24)
		offset += 4
		buf[offset] = byte(len(v.Depends))
		buf[offset+1] = byte(len(v.Depends) >> 8)
		offset += 2
		for _, v1 := range v.Depends {
			buf[offset] = byte(len(v1))
			buf[offset+1] = byte(len(v1) >> 8)
			offset += 2
			copy(buf[offset:], v1)
			offset += len(v1)
		}
		buf[offset] = byte(len(v.License))
		buf[offset+1] = byte(len(v.License) >> 8)
		offset += 2
		copy(buf[offset:], v.License)
		offset += len(v.License)
		buf[offset] = byte(len(v.MD5))
		buf[offset+1] = byte(len(v.MD5) >> 8)
		offset += 2
		copy(buf[offset:], v.MD5)
		offset += len(v.MD5)
		buf[offset] = byte(len(v.Name))
		buf[offset+1] = byte(len(v.Name) >> 8)
		offset += 2
		copy(buf[offset:], v.Name)
		offset += len(v.Name)
		buf[offset] = byte(len(v.Sha256))
		buf[offset+1] = byte(len(v.Sha256) >> 8)
		offset += 2
		copy(buf[offset:], v.Sha256)
		offset += len(v.Sha256)
		buf[offset] = byte(v.Size)
		buf[offset+1] = byte(v.Size >> 8)
		buf[offset+2] = byte(v.Size >> 16)
		buf[offset+3] = byte(v.Size >> 24)
		offset += 4
		buf[offset] = byte(len(v.Subdir))
		buf[offset+1] = byte(len(v.Subdir) >> 8)
		offset += 2
		copy(buf[offset:], v.Subdir)
		offset += len(v.Subdir)
		buf[offset] = byte(v.Timestamp)
		buf[offset+1] = byte(v.Timestamp >> 8)
		buf[offset+2] = byte(v.Timestamp >> 16)
		buf[offset+3] = byte(v.Timestamp >> 24)
		buf[offset+4] = byte(v.Timestamp >> 32)
		buf[offset+5] = byte(v.Timestamp >> 40)
		buf[offset+6] = byte(v.Timestamp >> 48)
		buf[offset+7] = byte(v.Timestamp >> 56)
		offset += 8
		buf[offset] = byte(len(v.Version))
		buf[offset+1] = byte(len(v.Version) >> 8)
		offset += 2
		copy(buf[offset:], v.Version)
		offset += len(v.Version)
	}
	buf[offset] = byte(len(s.PackagesConda))
	buf[offset+1] = byte(len(s.PackagesConda) >> 8)
	offset += 2
	for _, v := range s.PackagesConda {
		buf[offset] = byte(len(v.Package.Build))
		buf[offset+1] = byte(len(v.Package.Build) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.Build)
		offset += len(v.Package.Build)
		buf[offset] = byte(v.Package.BuildNumber)
		buf[offset+1] = byte(v.Package.BuildNumber >> 8)
		buf[offset+2] = byte(v.Package.BuildNumber >> 16)
		buf[offset+3] = byte(v.Package.BuildNumber >> 24)
		offset += 4
		buf[offset] = byte(len(v.Package.Depends))
		buf[offset+1] = byte(len(v.Package.Depends) >> 8)
		offset += 2
		for _, v1 := range v.Package.Depends {
			buf[offset] = byte(len(v1))
			buf[offset+1] = byte(len(v1) >> 8)
			offset += 2
			copy(buf[offset:], v1)
			offset += len(v1)
		}
		buf[offset] = byte(len(v.Package.License))
		buf[offset+1] = byte(len(v.Package.License) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.License)
		offset += len(v.Package.License)
		buf[offset] = byte(len(v.Package.MD5))
		buf[offset+1] = byte(len(v.Package.MD5) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.MD5)
		offset += len(v.Package.MD5)
		buf[offset] = byte(len(v.Package.Name))
		buf[offset+1] = byte(len(v.Package.Name) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.Name)
		offset += len(v.Package.Name)
		buf[offset] = byte(len(v.Package.Sha256))
		buf[offset+1] = byte(len(v.Package.Sha256) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.Sha256)
		offset += len(v.Package.Sha256)
		buf[offset] = byte(v.Package.Size)
		buf[offset+1] = byte(v.Package.Size >> 8)
		buf[offset+2] = byte(v.Package.Size >> 16)
		buf[offset+3] = byte(v.Package.Size >> 24)
		offset += 4
		buf[offset] = byte(len(v.Package.Subdir))
		buf[offset+1] = byte(len(v.Package.Subdir) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.Subdir)
		offset += len(v.Package.Subdir)
		buf[offset] = byte(v.Package.Timestamp)
		buf[offset+1] = byte(v.Package.Timestamp >> 8)
		buf[offset+2] = byte(v.Package.Timestamp >> 16)
		buf[offset+3] = byte(v.Package.Timestamp >> 24)
		buf[offset+4] = byte(v.Package.Timestamp >> 32)
		buf[offset+5] = byte(v.Package.Timestamp >> 40)
		buf[offset+6] = byte(v.Package.Timestamp >> 48)
		buf[offset+7] = byte(v.Package.Timestamp >> 56)
		offset += 8
		buf[offset] = byte(len(v.Package.Version))
		buf[offset+1] = byte(len(v.Package.Version) >> 8)
		offset += 2
		copy(buf[offset:], v.Package.Version)
		offset += len(v.Package.Version)
		buf[offset] = byte(len(v.Constrains))
		buf[offset+1] = byte(len(v.Constrains) >> 8)
		offset += 2
		for _, v1 := range v.Constrains {
			buf[offset] = byte(len(v1))
			buf[offset+1] = byte(len(v1) >> 8)
			offset += 2
			copy(buf[offset:], v1)
			offset += len(v1)
		}
		buf[offset] = byte(len(v.LegacyBz2Md5))
		buf[offset+1] = byte(len(v.LegacyBz2Md5) >> 8)
		offset += 2
		copy(buf[offset:], v.LegacyBz2Md5)
		offset += len(v.LegacyBz2Md5)
		buf[offset] = byte(len(v.LicenseFamily))
		buf[offset+1] = byte(len(v.LicenseFamily) >> 8)
		offset += 2
		copy(buf[offset:], v.LicenseFamily)
		offset += len(v.LicenseFamily)
	}
	buf[offset] = byte(len(s.Removed))
	buf[offset+1] = byte(len(s.Removed) >> 8)
	offset += 2
	for _, v := range s.Removed {
		buf[offset] = byte(len(v))
		buf[offset+1] = byte(len(v) >> 8)
		offset += 2
		copy(buf[offset:], v)
		offset += len(v)
	}
	buf[offset] = byte(s.RepoDataVersion)
	buf[offset+1] = byte(s.RepoDataVersion >> 8)
	buf[offset+2] = byte(s.RepoDataVersion >> 16)
	buf[offset+3] = byte(s.RepoDataVersion >> 24)
	offset += 4
	return w.Write(buf)
}

func (s *CondaRepoData) ReadFrom(r io.Reader) error {
	buf := make([]byte, 8)
	var size uint16
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	strBuf_0 := make([]byte, size)
	r.Read(strBuf_0)
	s.Info.Subdir = *(*string)(unsafe.Pointer(&strBuf_0))
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Packages = make([]Package, size)
	si := int(size)
	for i := 0; i < si; i++ {
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_1 := make([]byte, size)
		r.Read(strBuf_1)
		s.Packages[i].Build = *(*string)(unsafe.Pointer(&strBuf_1))
		r.Read(buf[:4])
		s.Packages[i].BuildNumber = uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) | (uint32(buf[3]) << 24)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		s.Packages[i].Depends = make([]string, size)
		si1 := int(size)
		for i1 := 0; i1 < si1; i1++ {
			r.Read(buf[:2])
			size = uint16(buf[0]) | (uint16(buf[1]) << 8)
			strBuf_2 := make([]byte, size)
			r.Read(strBuf_2)
			s.Packages[i].Depends[i1] = *(*string)(unsafe.Pointer(&strBuf_2))
		}
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_3 := make([]byte, size)
		r.Read(strBuf_3)
		s.Packages[i].License = *(*string)(unsafe.Pointer(&strBuf_3))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_4 := make([]byte, size)
		r.Read(strBuf_4)
		s.Packages[i].MD5 = *(*string)(unsafe.Pointer(&strBuf_4))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_5 := make([]byte, size)
		r.Read(strBuf_5)
		s.Packages[i].Name = *(*string)(unsafe.Pointer(&strBuf_5))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_6 := make([]byte, size)
		r.Read(strBuf_6)
		s.Packages[i].Sha256 = *(*string)(unsafe.Pointer(&strBuf_6))
		r.Read(buf[:4])
		s.Packages[i].Size = uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) | (uint32(buf[3]) << 24)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_7 := make([]byte, size)
		r.Read(strBuf_7)
		s.Packages[i].Subdir = *(*string)(unsafe.Pointer(&strBuf_7))
		r.Read(buf[:8])
		s.Packages[i].Timestamp = uint64(buf[0]) | (uint64(buf[1]) << 8) | (uint64(buf[2]) << 16) | (uint64(buf[3]) << 24) | (uint64(buf[4]) << 32) | (uint64(buf[5]) << 40) | (uint64(buf[6]) << 48) | (uint64(buf[7]) << 56)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_8 := make([]byte, size)
		r.Read(strBuf_8)
		s.Packages[i].Version = *(*string)(unsafe.Pointer(&strBuf_8))
	}
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.PackagesConda = make([]PackageConda, size)
	si2 := int(size)
	for i2 := 0; i2 < si2; i2++ {
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_9 := make([]byte, size)
		r.Read(strBuf_9)
		s.PackagesConda[i2].Package.Build = *(*string)(unsafe.Pointer(&strBuf_9))
		r.Read(buf[:4])
		s.PackagesConda[i2].Package.BuildNumber = uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) | (uint32(buf[3]) << 24)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		s.PackagesConda[i2].Package.Depends = make([]string, size)
		si3 := int(size)
		for i3 := 0; i3 < si3; i3++ {
			r.Read(buf[:2])
			size = uint16(buf[0]) | (uint16(buf[1]) << 8)
			strBuf_10 := make([]byte, size)
			r.Read(strBuf_10)
			s.PackagesConda[i2].Package.Depends[i3] = *(*string)(unsafe.Pointer(&strBuf_10))
		}
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_11 := make([]byte, size)
		r.Read(strBuf_11)
		s.PackagesConda[i2].Package.License = *(*string)(unsafe.Pointer(&strBuf_11))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_12 := make([]byte, size)
		r.Read(strBuf_12)
		s.PackagesConda[i2].Package.MD5 = *(*string)(unsafe.Pointer(&strBuf_12))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_13 := make([]byte, size)
		r.Read(strBuf_13)
		s.PackagesConda[i2].Package.Name = *(*string)(unsafe.Pointer(&strBuf_13))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_14 := make([]byte, size)
		r.Read(strBuf_14)
		s.PackagesConda[i2].Package.Sha256 = *(*string)(unsafe.Pointer(&strBuf_14))
		r.Read(buf[:4])
		s.PackagesConda[i2].Package.Size = uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) | (uint32(buf[3]) << 24)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_15 := make([]byte, size)
		r.Read(strBuf_15)
		s.PackagesConda[i2].Package.Subdir = *(*string)(unsafe.Pointer(&strBuf_15))
		r.Read(buf[:8])
		s.PackagesConda[i2].Package.Timestamp = uint64(buf[0]) | (uint64(buf[1]) << 8) | (uint64(buf[2]) << 16) | (uint64(buf[3]) << 24) | (uint64(buf[4]) << 32) | (uint64(buf[5]) << 40) | (uint64(buf[6]) << 48) | (uint64(buf[7]) << 56)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_16 := make([]byte, size)
		r.Read(strBuf_16)
		s.PackagesConda[i2].Package.Version = *(*string)(unsafe.Pointer(&strBuf_16))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		s.PackagesConda[i2].Constrains = make([]string, size)
		si4 := int(size)
		for i4 := 0; i4 < si4; i4++ {
			r.Read(buf[:2])
			size = uint16(buf[0]) | (uint16(buf[1]) << 8)
			strBuf_17 := make([]byte, size)
			r.Read(strBuf_17)
			s.PackagesConda[i2].Constrains[i4] = *(*string)(unsafe.Pointer(&strBuf_17))
		}
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_18 := make([]byte, size)
		r.Read(strBuf_18)
		s.PackagesConda[i2].LegacyBz2Md5 = *(*string)(unsafe.Pointer(&strBuf_18))
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_19 := make([]byte, size)
		r.Read(strBuf_19)
		s.PackagesConda[i2].LicenseFamily = *(*string)(unsafe.Pointer(&strBuf_19))
	}
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Removed = make([]string, size)
	si5 := int(size)
	for i5 := 0; i5 < si5; i5++ {
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		strBuf_20 := make([]byte, size)
		r.Read(strBuf_20)
		s.Removed[i5] = *(*string)(unsafe.Pointer(&strBuf_20))
	}
	r.Read(buf[:4])
	s.RepoDataVersion = uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) | (uint32(buf[3]) << 24)
	return nil
}
