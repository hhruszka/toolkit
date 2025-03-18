package testengine

var defaultPatterns = patternsV1

var patternsV1 = `
- pattern:
  name: Database Credentials
  regex: "dbUser=.*;dbPass=.*;"
  confidence: high
- pattern:
  name: SSH Keys
  regex: "ssh-rsa [A-Za-z0-9+/=]{100,}"
  confidence: high
- pattern:
  name: Certificates
  regex: "-----BEGIN CERTIFICATE-----[\\s\\S]+?-----END CERTIFICATE-----"
  confidence: high
- pattern:
  name: Passwords
  regex: "password=.*;"
  confidence: medium
- pattern:
  name: RSA private key
  regex: "-----BEGIN OPENSSH PRIVATE KEY-----"
  confidence: high
- pattern:
  name: RSA private key
  regex: "-----BEGIN RSA PRIVATE KEY-----"
  confidence: high
- pattern:
  name: SSH (DSA) private key
  regex: "-----BEGIN DSA PRIVATE KEY-----"
  confidence: high
- pattern:
  name: SSH (EC) private key
  regex: "-----BEGIN EC PRIVATE KEY-----"
  confidence: high
- pattern:
  name: PGP private key block
  regex: "-----BEGIN PGP PRIVATE KEY BLOCK-----"
  confidence: high
- pattern:
  name: Amazon MWS Auth Token
  regex: "amzn\\.mws\\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
  confidence: high
- pattern:
  name: AWS AppSync GraphQL Key
  regex: "da2-[a-z0-9]{26}"
  confidence: high
- pattern:
  name: GitHub
  regex: '[gG][iI][tT][hH][uU][bB].*[''|"][0-9a-zA-Z]{35,40}[''|"]'
  confidence: high
- pattern:
  name: Generic API Key
  regex: '[aA][pP][iI]_?[kK][eE][yY].*[''|"][0-9a-zA-Z]{32,45}[''|"]'
  confidence: high
- pattern:
  name: Generic Secret
  regex: '[sS][eE][cC][rR][eE][tT].*[''|"][0-9a-zA-Z]{32,45}[''|"]'
  confidence: high
- pattern:
  name: Google API Key
  regex: "AIza[0-9A-Za-z\\-_]{35}"
  confidence: high
- pattern:
  name: Google Cloud Platform API Key
  regex: "AIza[0-9A-Za-z\\-_]{35}"
  confidence: high
- pattern:
  name: Google Cloud Platform OAuth
  regex: "[0-9]+-[0-9A-Za-z_]{32}\\.apps\\.googleusercontent\\.com"
  confidence: high
#- pattern:
#  name: Google (GCP) Service-account
#  regex: '"type": "service_account"'
#  confidence: high
- pattern:
  name: Google OAuth Access Token
  regex: "ya29\\.[0-9A-Za-z\\-_]+"
  confidence: high
- pattern:
  name: Password in URL
  regex: "[a-zA-Z]{3,10}://[^/\\s:@]{3,20}:[^/\\s:@]{3,20}@.{1,100}[\"'\\s]"
  confidence: high
- pattern:
  name: Credentials
  regex: "[a-zA-Z0-9]{6,30}:[a-zA-Z0-9@#$%^&+=]{6,30}"
  confidence: high
- pattern:
  name: SSH Keys
  regex: "-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----\n([\\s\\S]*?)\n-----END (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----"
  confidence: high
- pattern:
  name: AWS Access Key ID
  regex: "AKIA[0-9A-Z]{16}"
  confidence: high
#- pattern:
#  name: AWS Secret Access Key
#  regex: "[A-Za-z0-9/+=]{40}"
#  confidence: high
- pattern:
  name: GCP Service Account Key
  regex: "[A-Za-z0-9\\-_]{20,60}\\.json"
  confidence: high
- pattern:
  name: Service Account Keys
  regex: "[A-Za-z0-9\\-]{20,60}\\.json"
  confidence: high
- pattern:
  name: TLS/SSL Certificates
  regex: "-----BEGIN CERTIFICATE-----\n([\\s\\S]*?)\n-----END CERTIFICATE-----"
  confidence: high
#- pattern:
#  name: Secret Keys for HMAC
#  regex: "[A-Za-z0-9+/=]{32,64}"
#  confidence: medium
# - pattern:
#  name: Environment-specific Configurations
#  regex: "[A-Z0-9_]{1,50}=[^\n]{1,100}"
#  confidence: high
- pattern:
  name: OAuth Tokens
  regex: "ya29\\.[0-9A-Za-z-_]+"
  confidence: high
- pattern:
  name: JWT Tokens
  regex: "eyJ[A-Za-z0-9-_=]+\\.eyJ[A-Za-z0-9-_=]+\\.?[A-Za-z0-9-_.+/=]*$"
  confidence: high
- pattern:
  name: Webhook Secrets
  regex: "whsec_[0-9A-Za-z]{32}"
  confidence: high
- pattern:
  name: API Keys
  regex: "\b[A-Za-z0-9]{35,45}\b"
  confidence: high
- pattern:
  name: IDs
  regex: "[A-Z]{3,}[-_]?ID=[A-Za-z0-9-]+"
  confidence: medium
`

var sshKeysPattern = `-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----\n([\\s\\S]*?)\n-----END (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----`
