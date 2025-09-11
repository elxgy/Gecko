// Commitlint configuration for Gecko CLI Editor
// Enforces conventional commit message format

module.exports = {
  extends: ['@commitlint/config-conventional'],
  
  // Custom rules for the Gecko project
  rules: {
    // Type enum - allowed commit types
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation changes
        'style',    // Code style changes (formatting, etc.)
        'refactor', // Code refactoring
        'perf',     // Performance improvements
        'test',     // Adding or updating tests
        'build',    // Build system or external dependencies
        'ci',       // CI/CD changes
        'chore',    // Maintenance tasks
        'revert',   // Reverting previous commits
        'security', // Security fixes
        'deps',     // Dependency updates
        'config',   // Configuration changes
        'ui',       // UI/UX improvements
        'a11y',     // Accessibility improvements
        'i18n',     // Internationalization
        'dx',       // Developer experience improvements
        'release'   // Release commits
      ]
    ],
    
    // Subject case - enforce lowercase
    'subject-case': [2, 'always', 'lower-case'],
    
    // Subject length - max 72 characters for better git log readability
    'subject-max-length': [2, 'always', 72],
    
    // Subject minimum length
    'subject-min-length': [2, 'always', 3],
    
    // Subject should not end with period
    'subject-full-stop': [2, 'never', '.'],
    
    // Body max line length
    'body-max-line-length': [2, 'always', 100],
    
    // Footer max line length
    'footer-max-line-length': [2, 'always', 100],
    
    // Header max length (type + scope + subject)
    'header-max-length': [2, 'always', 100],
    
    // Scope case - enforce lowercase
    'scope-case': [2, 'always', 'lower-case'],
    
    // Allowed scopes for the Gecko project
    'scope-enum': [
      2,
      'always',
      [
        // Core components
        'core',
        'editor',
        'buffer',
        'syntax',
        'ui',
        'model',
        'view',
        'controller',
        
        // Features
        'search',
        'selection',
        'clipboard',
        'file',
        'keybinds',
        'commands',
        'themes',
        'config',
        
        // Technical areas
        'performance',
        'memory',
        'rendering',
        'input',
        'output',
        'terminal',
        
        // Development
        'build',
        'test',
        'ci',
        'docs',
        'deps',
        'tools',
        'scripts',
        
        // LSP related (future)
        'lsp',
        'completion',
        'diagnostics',
        'hover',
        'definition',
        
        // Platform specific
        'linux',
        'macos',
        'windows',
        'cross-platform',
        
        // Security
        'security',
        'auth',
        'permissions',
        
        // API
        'api',
        'interface',
        'protocol'
      ]
    ],
    
    // Type case - enforce lowercase
    'type-case': [2, 'always', 'lower-case'],
    
    // Type should not be empty
    'type-empty': [2, 'never'],
    
    // Subject should not be empty
    'subject-empty': [2, 'never']
  },
  
  // Custom prompt configuration for interactive commits
  prompt: {
    questions: {
      type: {
        description: "Select the type of change that you're committing:",
        enum: {
          feat: {
            description: 'A new feature',
            title: 'Features',
            emoji: '✨'
          },
          fix: {
            description: 'A bug fix',
            title: 'Bug Fixes',
            emoji: '🐛'
          },
          docs: {
            description: 'Documentation only changes',
            title: 'Documentation',
            emoji: '📚'
          },
          style: {
            description: 'Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)',
            title: 'Styles',
            emoji: '💎'
          },
          refactor: {
            description: 'A code change that neither fixes a bug nor adds a feature',
            title: 'Code Refactoring',
            emoji: '📦'
          },
          perf: {
            description: 'A code change that improves performance',
            title: 'Performance Improvements',
            emoji: '🚀'
          },
          test: {
            description: 'Adding missing tests or correcting existing tests',
            title: 'Tests',
            emoji: '🚨'
          },
          build: {
            description: 'Changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)',
            title: 'Builds',
            emoji: '🛠'
          },
          ci: {
            description: 'Changes to our CI configuration files and scripts (example scopes: Travis, Circle, BrowserStack, SauceLabs)',
            title: 'Continuous Integrations',
            emoji: '⚙️'
          },
          chore: {
            description: "Other changes that don't modify src or test files",
            title: 'Chores',
            emoji: '♻️'
          },
          revert: {
            description: 'Reverts a previous commit',
            title: 'Reverts',
            emoji: '🗑'
          },
          security: {
            description: 'Security improvements or fixes',
            title: 'Security',
            emoji: '🔒'
          },
          deps: {
            description: 'Dependency updates',
            title: 'Dependencies',
            emoji: '📦'
          },
          ui: {
            description: 'UI/UX improvements',
            title: 'User Interface',
            emoji: '🎨'
          }
        }
      },
      scope: {
        description: 'What is the scope of this change (e.g. component or file name)'
      },
      subject: {
        description: 'Write a short, imperative tense description of the change'
      },
      body: {
        description: 'Provide a longer description of the change'
      },
      isBreaking: {
        description: 'Are there any breaking changes?'
      },
      breakingBody: {
        description: 'A BREAKING CHANGE commit requires a body. Please enter a longer description of the commit itself'
      },
      breaking: {
        description: 'Describe the breaking changes'
      },
      isIssueAffected: {
        description: 'Does this change affect any open issues?'
      },
      issuesBody: {
        description: 'If issues are closed, the commit requires a body. Please enter a longer description of the commit itself'
      },
      issues: {
        description: 'Add issue references (e.g. "fix #123", "re #123".)'
      }
    }
  },
  
  // Help URL for commit message format
  helpUrl: 'https://github.com/conventional-changelog/commitlint/#what-is-commitlint',
  
  // Default ignore patterns
  ignores: [
    // Ignore merge commits
    (commit) => commit.includes('Merge'),
    // Ignore revert commits (they have their own format)
    (commit) => commit.includes('Revert'),
    // Ignore release commits from automated tools
    (commit) => commit.includes('Release'),
    // Ignore WIP commits during development
    (commit) => commit.toLowerCase().includes('wip')
  ],
  
  // Default severity level
  defaultIgnores: true,
  
  // Custom formatter for better error messages
  formatter: '@commitlint/format'
};