package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.2"

func PrintVersion() {
	fmt.Printf("Current stoppiracy version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
          __                        _                          
   _____ / /_ ____   ____   ____   (_)_____ ____ _ _____ __  __
  / ___// __// __ \ / __ \ / __ \ / // ___// __  // ___// / / /
 (__  )/ /_ / /_/ // /_/ // /_/ // // /   / /_/ // /__ / /_/ / 
/____/ \__/ \____// .___// .___//_//_/    \__,_/ \___/ \__, /  
                 /_/    /_/                           /____/
`
	fmt.Printf("%s\n%60s\n\n", banner, "Current stoppiracy version "+version)
}
