package service

import (
	"net"

	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

type GeoIPService struct {
	reader *geoip2.Reader
	logger *zap.Logger
}

func NewGeoIPService(databasePath string, logger *zap.Logger) (*GeoIPService, error) {
	reader, err := geoip2.Open(databasePath)
	if err != nil {
		return nil, err
	}

	return &GeoIPService{
		reader: reader,
		logger: logger,
	}, nil
}

func (s *GeoIPService) GetCountry(ipAddress string) (string, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", nil
	}

	record, err := s.reader.City(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

func (s *GeoIPService) GetCity(ipAddress string) (string, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", nil
	}

	record, err := s.reader.City(ip)
	if err != nil {
		return "", err
	}

	return record.City.Names["en"], nil
}

func (s *GeoIPService) GetLocation(ipAddress string) (*GeoLocation, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return nil, nil
	}

	record, err := s.reader.City(ip)
	if err != nil {
		return nil, err
	}

	return &GeoLocation{
		Country:   record.Country.IsoCode,
		City:      record.City.Names["en"],
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
		Timezone:  record.Location.TimeZone,
	}, nil
}

func (s *GeoIPService) Close() {
	if s.reader != nil {
		s.reader.Close()
	}
}

type GeoLocation struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}
