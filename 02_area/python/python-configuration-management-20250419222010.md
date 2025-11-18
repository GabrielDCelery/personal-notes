---
title: Python configuration management
author: GaborZeller
date: 2025-04-19T22-20-10Z
tags:
draft: true
---

# Python configuration management

BAD PRACTICES:

1. Hardcoding sensitive values:

```python
# BAD: Never hardcode sensitive values
DATABASE_PASSWORD = "super_secret_password123"
API_KEY = "1234567890abcdef"
```

2. Unsafe direct access without validation:

```python
# BAD: No error handling or validation
api_key = os.environ['API_KEY']  # Will raise KeyError if not set
```

3. Using default values that might expose sensitive information:

```python
# BAD: Default values for sensitive data
database_url = os.environ.get('DATABASE_URL', 'postgresql://admin:password@localhost:5432/db')
```

4. Mixing configuration in code:

```python
# BAD: Mixing environment-specific config in code
if os.environ.get('ENV') == 'prod':
    API_URL = 'https://prod.api.com'
else:
    API_URL = 'http://localhost:8000'
```

GOOD PRACTICES:

1. Using environment variables with proper validation and error handling:

```python
import os
from typing import Optional
from dataclasses import dataclass
from pathlib import Path

@dataclass
class Config:
    """Application configuration."""

    @staticmethod
    def get_env_var(key: str, default: Optional[str] = None, required: bool = True) -> Optional[str]:
        """
        Retrieve and validate environment variables.

        Args:
            key: The environment variable key
            default: Optional default value if not set
            required: Whether the environment variable is required

        Returns:
            The environment variable value or default

        Raises:
            ValueError: If a required environment variable is missing
        """
        value = os.environ.get(key, default)
        if required and value is None:
            raise ValueError(f"Required environment variable '{key}' is not set")
        return value

    def __init__(self):
        # Required variables
        self.api_key = self.get_env_var("API_KEY", required=True)

        # Optional variables with sensible defaults
        self.debug_mode = self.get_env_var("DEBUG_MODE", default="False", required=False).lower() == "true"
        self.cache_dir = Path(self.get_env_var("CACHE_DIR", default="/tmp/cache", required=False))

        # Variables with validation
        log_level = self.get_env_var("LOG_LEVEL", default="INFO", required=False)
        if log_level not in ("DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"):
            raise ValueError(f"Invalid log level: {log_level}")
        self.log_level = log_level

# Usage
try:
    config = Config()
except ValueError as e:
    print(f"Configuration error: {e}")
    sys.exit(1)
```

2. Using environment file loading with python-dotenv:

```python
from dotenv import load_dotenv
import os
from pathlib import Path

def load_environment():
    """
    Load environment variables from .env file with fallbacks.
    """
    # Load from .env file if it exists
    env_path = Path('.env')
    load_dotenv(dotenv_path=env_path, override=True)

    # Load from .env.local if it exists (for local development)
    local_env_path = Path('.env.local')
    if local_env_path.exists():
        load_dotenv(dotenv_path=local_env_path, override=True)

# Usage
load_environment()
```

3. Using a centralized configuration management:

```python
from functools import lru_cache
from typing import Optional
from pydantic import BaseSettings, SecretStr

class Settings(BaseSettings):
    """
    Application settings management using pydantic.

    This provides validation, type checking, and secure handling of sensitive values.
    """
    # Sensitive values using SecretStr
    database_password: SecretStr
    api_key: SecretStr

    # Required values with validation
    database_host: str
    database_port: int

    # Optional values with defaults
    debug_mode: bool = False
    max_connections: int = 100

    class Config:
        env_file = '.env'
        env_file_encoding = 'utf-8'
        case_sensitive = True

@lru_cache()
def get_settings() -> Settings:
    """
    Get cached application settings.

    Returns:
        Settings instance with environment variables loaded
    """
    return Settings()

# Usage
settings = get_settings()
db_password = settings.database_password.get_secret_value()
```

Key Best Practices:

1. **Validation**: Always validate environment variables when loading them
2. **Type Safety**: Use type hints and validation to ensure correct types
3. **Security**: Use secure string handling for sensitive data
4. **Default Values**: Provide sensible defaults for non-sensitive configuration
5. **Error Handling**: Proper error handling for missing required variables
6. **Documentation**: Document all environment variables and their purpose
7. **Centralization**: Keep environment variable handling in one place
8. **Caching**: Cache configuration to avoid repeated environment variable lookups
9. **Separation**: Keep development and production configurations separate
10. **Tools**: Use established tools like python-dotenv or pydantic for robust handling

The good practices above provide several benefits:

- Type safety and validation
- Secure handling of sensitive data
- Clear error messages when configuration is missing
- Centralized configuration management
- Easy testing and development setup
- Clear separation of concerns
- Documentation of configuration options

Remember to never commit sensitive environment variables to version control. Always use:

- `.env.example` for documentation
- `.env` for local development (add to .gitignore)
- Proper secrets management for production environments
