package git

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// CopyService takes the name of a service to copy from a Source to a Destination.
// Source, Destination implement Walk() and are typically Repository objects.
//
// Only files under /services/[serviceName]/base/config/* are copied to the destination
//
// Returns the list of files that were copied, and possibly an error.
// Note this currently relies on the environment folder being in *both* the source and destination repos.
// sourceEnvironment is
// --env = destination environment
// todo: determine source before we call copy service, then make sure we use env as a true dest environment only
func CopyService(serviceName string, source Source, dest Destination, sourceEnvironment, destinationEnvironment string) ([]string, error) {
	// filePath defines the root folder for serviceName's config in the repository
	// the lookup is done for the source repository
	fmt.Println("in copy service")
	filePath := pathForServiceConfig(serviceName, sourceEnvironment)
	copied := []string{}
	err := source.Walk(filePath, func(prefix, name string) error {
		sourcePath := path.Join(prefix, name)
		destPath := pathForServiceConfig(name, destinationEnvironment)
		fmt.Printf("checking if valid for promotion, params: %s, %s, %s\n", serviceName, destPath, destinationEnvironment)
		if pathValidForPromotion(serviceName, destPath, destinationEnvironment) {
			fmt.Printf("Copying from path %s to path %s\n", sourcePath, destPath)
			err := dest.CopyFile(sourcePath, destPath)
			if err == nil {
				copied = append(copied, destPath)
			}
			return err
		}
		return nil
	})

	return copied, err
}

//  For a given serviceName, only files in environments/envName/services/serviceName/base/config/* are valid for promotion
func pathValidForPromotion(serviceName, filePath, environmentName string) bool {
	filterPath := filepath.Join(pathForServiceConfig(serviceName, environmentName), "base", "config")
	fmt.Printf("checking if %s starts with %s\n", filePath, filterPath)
	validPath := strings.HasPrefix(filePath, filterPath)
	return validPath
}

// pathForServiceConfig defines where in a 'gitops' repository the config for a given service should live.
func pathForServiceConfig(serviceName, environmentName string) string {
	// Strip environments if it was somehow provided
	if strings.Contains(environmentName, "environments") {
		environmentName = strings.Replace(environmentName, "environments", "", -1)
	}
	// Todo does this defeat the point of .Join? / isn't portable, but I want a leading slash (top-level dir) don't I
	pathForConfig := filepath.Join("/", "environments", environmentName, "services", serviceName)
	return pathForConfig
}
