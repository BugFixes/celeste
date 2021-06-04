package account

import (
  "github.com/bugfixes/celeste/internal/config"
)

type Account struct {
  Config config.Config
}

func NewAccount(c config.Config) *Account {
  return &Account{
    Config: c,
  }
}
