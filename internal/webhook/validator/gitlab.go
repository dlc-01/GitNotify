package validator

const sourceGitLab = "gitlab"

type GitLabValidator struct {
	secret string
}

func NewGitLabValidator(secret string) *GitLabValidator {
	return &GitLabValidator{secret: secret}
}

func (v *GitLabValidator) Source() string { return sourceGitLab }

func (v *GitLabValidator) Validate(payload []byte, headers map[string]string) error {
	if v.secret == "" {
		return wrap("Validate", sourceGitLab, ErrEmptySecret)
	}

	token, ok := headers["x-gitlab-token"]
	if !ok {
		return wrap("Validate", sourceGitLab, ErrMissingHeader)
	}

	if token != v.secret {
		return wrap("Validate", sourceGitLab, ErrInvalidToken)
	}

	return nil
}
