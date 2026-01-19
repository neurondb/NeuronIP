# Security Policy

## Supported Versions

We actively support the following versions of NeuronIP with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of NeuronIP seriously. If you discover a security vulnerability, please follow these steps:

### Please DO NOT:

- Open a public GitHub issue
- Discuss the vulnerability publicly until it has been addressed

### Please DO:

1. **Email us directly** at the security contact for the project (if available) or open a private security advisory on GitHub
2. **Provide detailed information** about the vulnerability:
   - Type of vulnerability (XSS, SQL injection, authentication bypass, etc.)
   - Affected components (backend, frontend, database, etc.)
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if you have one)

### What to Expect

- **Initial Response**: Within 48 hours, we will acknowledge receipt of your report
- **Status Updates**: We'll provide updates on the progress of the fix
- **Resolution Timeline**: We'll work to address critical vulnerabilities within 7 days, and high-severity issues within 30 days
- **Credit**: With your permission, we'll credit you in the security advisory and release notes

### Security Best Practices

When using NeuronIP in production:

1. **Keep dependencies updated**: Regularly update Go modules and npm packages
2. **Use environment variables**: Never commit secrets or API keys
3. **Enable HTTPS**: Always use HTTPS in production
4. **Review permissions**: Implement proper RBAC and access controls
5. **Monitor logs**: Regularly review application and access logs
6. **Database security**: Use strong passwords and enable SSL for database connections
7. **Network security**: Restrict access to services using firewalls and security groups

### Security Considerations

NeuronIP includes several security features:

- Authentication and authorization middleware
- CORS configuration
- Input validation
- SQL injection prevention (parameterized queries)
- Rate limiting
- Security headers

Always ensure these are properly configured in your deployment.

## Disclosure Policy

When a security vulnerability is reported:

1. We confirm the issue and determine affected versions
2. We develop a fix and test it thoroughly
3. We release the fix in a new version
4. We publish a security advisory with details and credits

We follow responsible disclosure practices to protect users while giving them time to update.

## Questions?

If you have questions about this security policy, please open a discussion or contact the maintainers.
