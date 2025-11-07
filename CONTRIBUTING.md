# Contributing to IPTV Platform

Thank you for your interest in contributing to IPTV Platform! This document provides guidelines and instructions for contributing.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

---

## 📜 Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We expect all contributors to:

- Be respectful and considerate
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker & Docker Compose
- MySQL 8.0
- Git

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/iptv-platform.git
cd iptv-platform
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/original/iptv-platform.git
```

---

## 💻 Development Setup

### 1. Install Dependencies

**Backend:**
```bash
cd microservices/streaming-gateway
go mod download
```

**Frontend:**
```bash
cd admin-dashboard
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your local settings
```

### 3. Start Development Environment

```bash
docker-compose -f docker-compose.dev.yml up -d
```

### 4. Run Migrations

```bash
./scripts/migrate.sh
./scripts/seed-database.sh
```

### 5. Start Development Servers

**Backend:**
```bash
cd microservices/streaming-gateway
go run cmd/server/main.go
```

**Frontend:**
```bash
cd admin-dashboard
npm run dev
```

---

## 🤝 How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title** - Brief description of the issue
- **Steps to reproduce** - Detailed steps to recreate the bug
- **Expected behavior** - What you expected to happen
- **Actual behavior** - What actually happened
- **Environment** - OS, browser, versions, etc.
- **Screenshots** - If applicable
- **Error logs** - Relevant error messages

**Bug Report Template:**

```markdown
## Bug Description
[Clear and concise description]

## Steps to Reproduce
1. Go to '...'
2. Click on '...'
3. See error

## Expected Behavior
[What should happen]

## Actual Behavior
[What actually happens]

## Environment
- OS: [e.g., Ubuntu 22.04]
- Browser: [e.g., Chrome 120]
- Version: [e.g., 1.0.0]

## Additional Context
[Any other relevant information]
```

### Suggesting Features

Feature requests are welcome! Please provide:

- **Use case** - Why is this feature needed?
- **Proposed solution** - How should it work?
- **Alternatives** - Other solutions you've considered
- **Priority** - Low, Medium, High, Critical

### Code Contributions

1. **Pick an issue** - Look for issues labeled `good first issue` or `help wanted`
2. **Create a branch** - Use descriptive branch names
3. **Make changes** - Follow coding standards
4. **Write tests** - Add tests for new features
5. **Update docs** - Document new features
6. **Submit PR** - Create a pull request

---

## 📝 Coding Standards

### Go (Backend)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golint` and `go vet`
- Write godoc comments for exported functions
- Keep functions small and focused
- Error handling is required

**Example:**

```go
// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id int64) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidUserID
    }
    
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}
```

### TypeScript/React (Frontend)

- Follow [Airbnb React/JSX Style Guide](https://github.com/airbnb/javascript/tree/master/react)
- Use TypeScript strict mode
- Use functional components with hooks
- Props should be typed with interfaces
- Use meaningful variable names
- Keep components small and reusable

**Example:**

```typescript
interface UserCardProps {
  user: User;
  onEdit: (id: number) => void;
  onDelete: (id: number) => void;
}

export const UserCard: React.FC<UserCardProps> = ({ user, onEdit, onDelete }) => {
  return (
    <div className="user-card">
      <h3>{user.name}</h3>
      <p>{user.email}</p>
      <button onClick={() => onEdit(user.id)}>Edit</button>
      <button onClick={() => onDelete(user.id)}>Delete</button>
    </div>
  );
};
```

### SQL

- Use meaningful table and column names
- Include comments for complex queries
- Use indexes appropriately
- Follow naming conventions

---

## 📝 Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Commit Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

### Examples

```
feat(auth): add JWT refresh token functionality

Implemented refresh token mechanism to allow users to stay
logged in without re-authenticating.

Closes #123
```

```
fix(streaming): resolve buffering issue on mobile

Fixed a race condition that caused buffering on mobile devices
when switching between qualities.

Fixes #456
```

---

## 🔄 Pull Request Process

### Before Submitting

1. **Sync with upstream:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests:**
   ```bash
   # Backend tests
   cd microservices/streaming-gateway
   go test ./...

   # Frontend tests
   cd admin-dashboard
   npm run test
   ```

3. **Run linters:**
   ```bash
   # Backend
   golint ./...
   go vet ./...

   # Frontend
   npm run lint
   ```

### Submitting PR

1. **Create descriptive PR title** - Follow commit convention
2. **Fill out PR template** - Provide all requested information
3. **Link related issues** - Use "Closes #123" or "Fixes #456"
4. **Request review** - Tag relevant reviewers
5. **Be responsive** - Address review comments promptly

### PR Template

```markdown
## Description
[Describe your changes]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## How Has This Been Tested?
[Describe testing approach]

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added where needed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests passing
- [ ] No new warnings
```

---

## 🧪 Testing

### Writing Tests

**Backend (Go):**

```go
func TestGetUserByID(t *testing.T) {
    // Setup
    service := NewUserService(mockRepo)
    
    // Test
    user, err := service.GetUserByID(1)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

**Frontend (TypeScript):**

```typescript
describe('UserCard', () => {
  it('renders user information', () => {
    const user = { id: 1, name: 'John', email: 'john@example.com' };
    render(<UserCard user={user} onEdit={() => {}} onDelete={() => {}} />);
    
    expect(screen.getByText('John')).toBeInTheDocument();
    expect(screen.getByText('john@example.com')).toBeInTheDocument();
  });
});
```

### Running Tests

```bash
# Run all tests
npm run test:all

# Run backend tests
cd microservices/streaming-gateway
go test ./... -v

# Run frontend tests
cd admin-dashboard
npm run test

# Run with coverage
go test ./... -cover
npm run test:coverage
```

---

## 📚 Documentation

### Code Documentation

- Add godoc comments for Go functions
- Add JSDoc comments for TypeScript functions
- Document complex algorithms
- Keep README.md updated

### API Documentation

- Document all API endpoints
- Include request/response examples
- Specify error codes
- Update Postman collection

---

## 🏷️ Issue Labels

- `bug` - Something isn't working
- `feature` - New feature request
- `documentation` - Documentation improvements
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `enhancement` - Improvement to existing feature
- `question` - Further information requested
- `wontfix` - This will not be worked on

---

## 🎯 Review Process

1. **Automated checks** - CI/CD must pass
2. **Code review** - At least one approval required
3. **Testing** - All tests must pass
4. **Documentation** - Docs must be updated
5. **Approval** - Maintainer approval needed
6. **Merge** - Squash and merge

---

## 💬 Communication

- **GitHub Issues** - Bug reports, feature requests
- **GitHub Discussions** - General questions, ideas
- **Discord** - Real-time chat with community
- **Email** - For private/security concerns

---

## 📞 Need Help?

- Check existing documentation
- Search closed issues
- Ask in GitHub Discussions
- Join our Discord server
- Email: developers@iptv.example.com

---

## 🙏 Thank You!

Your contributions make this project better for everyone. Thank you for taking the time to contribute!

---

*Last updated: January 2025*
