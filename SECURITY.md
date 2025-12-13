# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of abp-gen seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please do NOT:

- Open a public GitHub issue
- Discuss the vulnerability publicly
- Share the vulnerability with others until it has been resolved

### Please DO:

1. **Email us directly** at: [mohamedhabibwork@gmail.com](mailto:mohamedhabibwork@gmail.com)

2. **Include the following information:**
   - Type of vulnerability (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
   - Full paths of source file(s) related to the vulnerability
   - The location of the affected source code (tag/branch/commit or direct URL)
   - Step-by-step instructions to reproduce the issue
   - Proof-of-concept or exploit code (if possible)
   - Impact of the issue, including how an attacker might exploit the issue

3. **We will:**
   - Acknowledge receipt of your report within 48 hours
   - Provide an initial assessment within 7 days
   - Keep you informed of our progress
   - Notify you when the vulnerability has been fixed
   - Credit you in the security advisory (unless you prefer to remain anonymous)

### Security Response Timeline

- **48 hours**: Initial acknowledgment
- **7 days**: Initial assessment and severity classification
- **30 days**: Target fix release (depending on severity)
- **90 days**: Public disclosure (if not fixed earlier)

### Severity Classification

We use CVSS v3.0 to assess severity:

- **Critical (9.0-10.0)**: Immediate fix required
- **High (7.0-8.9)**: Fix within 30 days
- **Medium (4.0-6.9)**: Fix within 90 days
- **Low (0.1-3.9)**: Fix in next regular release

### Security Best Practices

When using abp-gen:

- Always use the latest stable version
- Review generated code before deploying to production
- Keep your Go toolchain and dependencies updated
- Use secure coding practices in generated code
- Validate all inputs in generated applications
- Follow ABP Framework security guidelines

### Known Security Considerations

- **Code Generation**: Generated code should be reviewed for security before use
- **File System Access**: The tool reads and writes files - ensure proper permissions
- **Template Injection**: Templates are executed with user input - validate schemas
- **Dependencies**: Keep dependencies updated to avoid known vulnerabilities

### Security Updates

Security updates will be:
- Released as patch versions (e.g., 1.0.1, 1.0.2)
- Documented in CHANGELOG.md
- Announced via GitHub releases
- Tagged with security advisories

### Credits

We appreciate responsible disclosure and will credit security researchers who help us improve the security of abp-gen.

Thank you for helping keep abp-gen and its users safe!

