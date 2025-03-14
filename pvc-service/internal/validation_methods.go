package internal

import (
	"log"
	"regexp"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// isValidKubernetesName validates if a given name adheres to Kubernetes naming conventions
var IsValidKubernetesName = func(name string) error {

	if name == "" {
		return status.Errorf(codes.InvalidArgument, "volume name cannot be empty")
	}

	// Kubernetes names must be a DNS-1123 subdomain
	// This regex checks for lowercase alphanumeric characters, '-' and '.' but not start or end with '-'
	// Max length of 253 characters
	const dns1123Regex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	if len(name) > 253 {
		return status.Errorf(codes.InvalidArgument, "invalid volume name: %s", name)
	}
	re := regexp.MustCompile(dns1123Regex)
	if !re.MatchString(name) {
		return status.Errorf(codes.InvalidArgument, "invalid volume name: %s", name)
	}
	return nil
}

// isValidSize validates if the given size is a positive integer
var IsValidSize = func(sizePtr *string) string {
	var size string
	if sizePtr == nil || *sizePtr == "" {
		size = "2"
	} else {
		size = *sizePtr
	}
	// Check if the size contains only digits using a regular expression
	match, _ := regexp.MatchString(`^\d+$`, size)
	if !match {
		return ""
	}
	// Check if the size is a valid positive integer
	log.Println(size)
	num, err := strconv.Atoi(size)
	if err != nil || num <= 0 {
		return ""
	}
	return size
}
