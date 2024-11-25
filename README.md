# MFA Email Client

A Go package that provides a simple and secure way to implement email-based Multi-Factor Authentication (MFA) in your applications.

## Features:
- Generates secure 6-digit MFA codes
- Sends MFA codes via email using SMTP
- Configurable code expiration
- Automatic code cleanup after expiration
- Thread-safe implementation using sync.Map
- Email validation
- Customizable email templates

## Installation:
```bash
go get github.com/yingwei123/mfa@v1.0.0
```

## Quick Start Example:
```go
package main

import (
    "context"
    "mfa/client"
    "time"
)

func main() {
    // Create a new MFA client
    mfaClient, err := client.CreateMFAClient(
        5*time.Minute,           // Code duration
        "your@email.com",        // SMTP client email
        "your-password",         // SMTP password
        "smtp.gmail.com",        // SMTP server
        587,                     // SMTP port
        "<div>Your code: {{.MfaCode}}</div>", // Email template
    )
    if err != nil {
        panic(err)
    }

    // Send MFA code
    err = mfaClient.SendMFAEmail(context.Background(), "user@example.com")
    if err != nil {
        panic(err)
    }

    // Verify MFA code
    err = mfaClient.VerifyMFA("user@example.com", "123456")
    if err != nil {
        panic(err)
    }
}
```

A more comprehensive example can be found in the [example](example/main.go) folder.

## Configuration:

MFA Client Options:
- codeDuration: Duration until the MFA code expires
- clientEmail: Email address used to send MFA codes
- clientPassword: Password for the email account
- smtpServer: SMTP server address
- smtpServerPort: SMTP server port
- mfaEmailTemplate: HTML template for the email (must include {{.MfaCode}})

Email Template Example:
```html
<div>Your MFA code is {{.MfaCode}}. This code will expire in 5 minutes.</div>
```

# Configuring Gmail SMTP Settings

To configure Gmail SMTP settings, follow these steps:

## For MFA-Enabled Accounts:
1. Go to your **Google Account's Security page**.
2. Find the **"App passwords"** section.
3. Create a new **App password**.
4. Use the following SMTP settings:

   - **SMTP server**: `smtp.gmail.com`
   - **Username**: Your complete Gmail address
   - **Password**: The App password you generated
   - **Port**: `587` (for TLS) or `465` (for SSL)
   - **Enable TLS/STARTTLS**: Yes

## For Non-MFA Accounts:
If you don't have MFA enabled, the password will simply be your Gmail account password. Use the same SMTP settings as above, substituting your Gmail password in place of the App password.

## Features in Detail:

### Code Generation:
- Generates cryptographically secure 6-digit codes
- Uses Go's crypto/rand package for secure random number generation

### Email Sending:
- Supports SMTP with TLS
- Validates email addresses
- Customizable email templates
- Context support for timeouts and cancellation

### Code Verification:
- Thread-safe code storage using sync.Map
- Automatic code expiration
- One-time use codes (deleted after verification)
- Protection against brute-force attempts

### Error Handling:
The package provides detailed error messages for common scenarios:
- Invalid email addresses
- SMTP configuration errors
- Expired MFA codes
- Invalid MFA codes
- Missing MFA codes
- Template parsing errors

### Best Practices:
1. Always use HTTPS when implementing MFA in web applications
2. Set appropriate code expiration times (recommended: 5-15 minutes)
3. Implement rate limiting for MFA code requests
4. Store SMTP credentials securely (use environment variables)
5. Handle errors appropriately in your application

## Contributing:
Contributions are welcome! Please feel free to submit a Pull Request.

## License:
This project is licensed under the MIT License - see the LICENSE file for details.
