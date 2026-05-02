env "local" {
  src = "file://schema.sql"
  
  dev = "postgres://postgres:postgres@localhost:5432/fms_dev?sslmode=disable"
  
  migration {
    dir = "file://."
  }
  
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

