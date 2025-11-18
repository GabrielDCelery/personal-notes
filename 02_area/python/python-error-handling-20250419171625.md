---
title: Python error handling
author: GaborZeller
date: 2025-04-19T17-16-25Z
tags:
draft: true
---

# Python error handling

# Best Practices for Python Error Handling

## 1. Be Specific with Exception Types

❌ Bad Practice:

```python
try:
    do_something()
except Exception as e:  # Too broad! Catches everything
    print(f"An error occurred: {e}")
```

✅ Good Practice:

```python
try:
    response = requests.get(url)
    response.raise_for_status()
except requests.RequestException as e:
    logger.error(f"HTTP request failed: {e}")
except json.JSONDecodeError as e:
    logger.error(f"Failed to parse JSON response: {e}")
```

## 2. Use Context Managers (with statements)

❌ Bad Practice:

```python
file = open("file.txt", "w")
try:
    file.write("some data")
finally:
    file.close()
```

✅ Good Practice:

```python
with open("file.txt", "w") as file:
    file.write("some data")
```

## 3. Proper Logging Instead of Print

❌ Bad Practice:

```python
try:
    process_data()
except ValueError as e:
    print(f"Error: {e}")  # Using print for errors
```

✅ Good Practice:

```python
try:
    process_data()
except ValueError as e:
    logger.error("Failed to process data", exc_info=True)
    # or
    logger.exception("Failed to process data")  # Automatically includes traceback
```

## 4. Custom Exceptions for Better Error Handling

✅ Good Practice:

```python
class ImageDownloadError(Exception):
    """Raised when an image download fails"""
    pass

class InvalidImageURLError(ImageDownloadError):
    """Raised when the image URL is invalid"""
    pass

async def download_image(session: ClientSession, image_url: str, download_dir: str) -> None:
    if not image_url.startswith(('http://', 'https://')):
        raise InvalidImageURLError(f"Invalid URL scheme: {image_url}")

    try:
        async with session.get(image_url) as response:
            if response.status != 200:
                raise ImageDownloadError(f"Failed to download image. Status: {response.status}")
            # ... rest of the code
    except aiohttp.ClientError as e:
        raise ImageDownloadError(f"Network error while downloading image: {e}")
```

## 5. Clean Resource Management

❌ Bad Practice:

```python
conn = database.connect()
try:
    data = conn.query("SELECT * FROM users")
    # What if an error occurs here?
conn.close()  # This might never be called!
```

✅ Good Practice:

```python
class DatabaseConnection:
    def __enter__(self):
        self.conn = database.connect()
        return self.conn

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.conn.close()

with DatabaseConnection() as conn:
    data = conn.query("SELECT * FROM users")
```

## 6. Don't Suppress Exceptions Without Good Reason

❌ Bad Practice:

```python
try:
    process_data()
except Exception:
    pass  # Silently ignoring errors is dangerous!
```

✅ Good Practice:

```python
try:
    process_data()
except ValueError as e:
    logger.warning(f"Invalid data format: {e}")
    # Handle the error appropriately
    return default_value
```

## 7. Proper Exception Chaining

❌ Bad Practice:

```python
try:
    process_data()
except ValueError as e:
    raise RuntimeError("Processing failed")  # Original error context is lost
```

✅ Good Practice:

```python
try:
    process_data()
except ValueError as e:
    raise RuntimeError("Processing failed") from e  # Maintains error context
```

## 8. Use finally for Cleanup

✅ Good Practice:

```python
lock = threading.Lock()
try:
    lock.acquire()
    # do something with shared resource
finally:
    lock.release()  # Always release the lock, even if an error occurs
```

## 9. Avoid Catching and Re-raising the Same Exception

❌ Bad Practice:

```python
try:
    do_something()
except ValueError as e:
    raise ValueError(str(e))  # Pointless! Just let it propagate
```

✅ Good Practice:

```python
try:
    do_something()
except ValueError as e:
    # Only catch if you're adding value
    logger.error(f"Invalid value in do_something: {e}")
    raise  # If you must re-raise, use bare 'raise'
```

## 10. Use else Clause When No Exception Means Success

✅ Good Practice:

```python
try:
    data = process_file()
except IOError as e:
    logger.error("Failed to process file", exc_info=True)
    return None
else:
    # This only runs if no exception occurred
    logger.info("File processed successfully")
    return data
finally:
    # This always runs
    cleanup_resources()
```

Let's apply these principles to your image scraper code:

```python
async def download_image(
    session: ClientSession, image_url: str, download_dir: str
) -> None:
    """Download a single image from the given URL.

    Args:
        session: The aiohttp client session
        image_url: URL of the image to download
        download_dir: Directory to save the image

    Raises:
        ImageDownloadError: If the download fails
        InvalidImageURLError: If the URL is invalid
    """
    try:
        file_name = image_url.split("/")[-1]
        file_path = os.path.join(download_dir, file_name)

        if os.path.isfile(file_path):
            logger.debug(f"Skipping existing file: {file_name}")
            return

        if not image_url.startswith(('http://', 'https://')):
            raise InvalidImageURLError(f"Invalid URL scheme: {image_url}")

        async with session.get(image_url) as response:
            if response.status != 200:
                raise ImageDownloadError(
                    f"Failed to download {file_name}. Status: {response.status}"
                )

            logger.debug(f"Downloading {file_name}")
            content = await response.read()

            # Use context manager for file operations
            with open(file_path, "wb") as f:
                f.write(content)

            logger.info(f"Successfully downloaded: {file_name}")

    except aiohttp.ClientError as e:
        logger.error(f"Network error downloading {image_url}: {e}", exc_info=True)
        raise ImageDownloadError(f"Failed to download {image_url}") from e
    except OSError as e:
        logger.error(f"File system error saving {file_path}: {e}", exc_info=True)
        raise ImageDownloadError(f"Failed to save {file_name}") from e
```

Key points in this improved version:

1. Uses custom exceptions for better error handling
2. Properly logs errors with context
3. Uses context managers for file operations
4. Includes docstring with raised exceptions
5. Maintains exception context with `raise ... from`
6. Has specific exception handling for different error types
7. Provides meaningful error messages
8. Uses proper logging levels (debug for verbose info, error for errors)

Remember:

- Always log exceptions with enough context to debug issues
- Use custom exceptions when it adds value
- Don't catch exceptions unless you can handle them meaningfully
- Use context managers for resource cleanup
- Be specific about which exceptions you catch
- Include proper documentation about possible exceptions
