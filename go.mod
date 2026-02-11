module userService

go 1.25.4

require github.com/BurntSushi/toml v1.6.0

require github.com/lib/pq v1.11.1

require github.com/alexedwards/scs/v2 v2.9.0

require github.com/google/uuid v1.6.0 // indirect

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	golang.org/x/crypto v0.48.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
