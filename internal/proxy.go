package internal

import (
	"github.com/spf13/afero"
	"log"
	"time"
)

func RunProxy(duration time.Duration, certsPath, nginxSettingsPath, kubeNamespace string) {
	appFs := afero.NewOsFs()
	nginx := NewNginx(appFs, certsPath, nginxSettingsPath)
	kubeApi := NewKubeApi(kubeNamespace)

	for {
		SynchronizeServers(nginx, kubeApi, nginx.certsPath)

		time.Sleep(duration)
	}
}

func SynchronizeServers(nginx ProxyServer, kubeApi ServerSettingsStorage, certsPath string) {
	localServers, err := nginx.ReadServers()
	if err != nil {
		log.Printf("Failed to read local Nginx configs: %v", err)
		return
	}
	remoteServers, err := kubeApi.ReadServers()
	if err != nil {
		log.Printf("Failed to read Kubernetes configs: %v", err)
		return
	}

	for _, s := range remoteServers {
		s.FillTemplate(certsPath)
	}

	changedServers := SelectChangedServers(localServers, remoteServers)
	removedServers := SelectRemovedServers(localServers, remoteServers)

	if len(removedServers) > 0 {
		for _, s := range removedServers {
			if err := nginx.RemoveServer(s); err != nil {
				log.Printf("Failed to remove server config for domain %s: %v", s.Domain, err)
				return
			} else {
				log.Printf("Removed server config for domain %s", s.Domain)
			}
		}
	}

	if len(changedServers) > 0 {
		for _, s := range changedServers {
			if err = nginx.WriteServer(s); err != nil {
				log.Printf("Failed to write server config for domain %s: %v", s.Domain, err)
				return
			} else {
				log.Printf("Write server config for domain %s", s.Domain)
			}
		}
	}

	err = nginx.Restart()
	if err != nil {
		log.Printf("Failed update nginx: %v", err)
	} else {
		log.Println("Nginx updated")
	}
}
