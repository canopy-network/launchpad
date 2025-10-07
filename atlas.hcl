// Atlas configuration file for Launchpad database migrations

variable "db_url" {
  type = string
  default = "postgres://launchpad:launchpad123@localhost:5432/launchpad?search_path=public&sslmode=disable"
}

variable "dev_db_url" {
  type = string
  default = "docker://postgres/15/dev?search_path=public"
}

env "local" {
  // Local development environment
  src = "file://schema.sql"
  url = var.db_url
  dev = var.dev_db_url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "docker" {
  // Docker compose environment
  src = "file://schema.sql"
  url = "postgres://launchpad:launchpad123@localhost:5432/launchpad?search_path=public&sslmode=disable"
  dev = var.dev_db_url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

// Production environment template (customize as needed)
env "prod" {
  src = "file://schema.sql"
  url = getenv("DATABASE_URL")
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}