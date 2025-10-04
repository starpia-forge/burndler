import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ContainerVersionForm from './ContainerVersionForm';
import { ContainerVersion } from '../../types/container';

// Mock i18n
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: any) => {
      if (key === 'containers:versionCharacterCount') {
        return `${options?.count}/${options?.max} characters`;
      }
      if (key === 'containers:composeCharacterCount') {
        return `${options?.count}/${options?.max} characters`;
      }
      return key;
    },
  }),
}));

// Mock ConfigurationSelector
vi.mock('./ConfigurationSelector', () => ({
  ConfigurationSelector: ({ value, onChange }: any) => (
    <div data-testid="configuration-selector">
      <select
        aria-label="configuration"
        value={value || ''}
        onChange={(e) => onChange(e.target.value ? parseInt(e.target.value) : null)}
      >
        <option value="">No Configuration</option>
        <option value="1">default</option>
        <option value="2">production</option>
      </select>
    </div>
  ),
}));

const mockInitialVersion: ContainerVersion = {
  id: 1,
  container_id: 1,
  version: 'v1.0.0',
  compose_content: 'version: "3.8"\nservices:\n  app:\n    image: nginx:alpine',
  published: false,
  configuration_id: 1,
  variables: { ENV: 'test' },
  resource_paths: ['/path/to/resource'],
  dependencies: { service: 'v1.0.0' },
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
};

describe('ContainerVersionForm', () => {
  const mockOnSubmit = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('create mode', () => {
    it('should render form in create mode by default', () => {
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      expect(screen.getByLabelText(/containers:versionNumber/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/containers:dockerComposeYaml/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /common:cancel/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /containers:createVersion/i })).toBeInTheDocument();
    });

    it('should have editable version field in create mode', () => {
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      expect(versionInput).not.toBeDisabled();
    });

    it('should not show configuration selector in create mode without initialData', () => {
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      expect(screen.queryByTestId('configuration-selector')).not.toBeInTheDocument();
    });

    it('should validate version field is required', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:versionRequired/i)).toBeInTheDocument();
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should validate version format', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      await user.type(versionInput, 'invalid version!');

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:versionInvalidFormat/i)).toBeInTheDocument();
    });

    it('should validate version max length', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      await user.type(versionInput, 'a'.repeat(51));

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:versionMaxLength/i)).toBeInTheDocument();
    });

    it('should show character count for version field', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      await user.type(versionInput, 'v1.0.0');

      expect(screen.getByText('6/50 characters')).toBeInTheDocument();
    });

    it('should submit valid create data', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      const composeInput = screen.getByLabelText(/containers:dockerComposeYaml/i);

      await user.type(versionInput, 'v1.0.0');
      await user.clear(composeInput);
      await user.type(composeInput, 'version: "3.8"');

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(mockOnSubmit).toHaveBeenCalledWith({
        version: 'v1.0.0',
        compose: 'version: "3.8"',
        variables: {},
        resource_paths: [],
        dependencies: {},
      });
    });
  });

  describe('edit mode', () => {
    it('should render form in edit mode with initial data', () => {
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      expect(screen.getByRole('button', { name: /containers:updateVersion/i })).toBeInTheDocument();
      expect(screen.getByDisplayValue('v1.0.0')).toBeInTheDocument();
    });

    it('should disable version field in edit mode', () => {
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      expect(versionInput).toBeDisabled();
      expect(screen.getByText(/version number cannot be changed/i)).toBeInTheDocument();
    });

    it('should show configuration selector in edit mode', () => {
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      expect(screen.getByTestId('configuration-selector')).toBeInTheDocument();
    });

    it('should not validate version field in edit mode', async () => {
      const user = userEvent.setup();
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      const submitButton = screen.getByRole('button', { name: /containers:updateVersion/i });
      await user.click(submitButton);

      // Should not show version validation errors
      expect(screen.queryByText(/containers:versionRequired/i)).not.toBeInTheDocument();
    });

    it('should submit valid update data', async () => {
      const user = userEvent.setup();
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      const composeInput = screen.getByLabelText(/containers:dockerComposeYaml/i);
      await user.clear(composeInput);
      await user.type(composeInput, 'version: "3.9"');

      const submitButton = screen.getByRole('button', { name: /containers:updateVersion/i });
      await user.click(submitButton);

      expect(mockOnSubmit).toHaveBeenCalledWith({
        compose: 'version: "3.9"',
        variables: { ENV: 'test' },
        resource_paths: ['/path/to/resource'],
        dependencies: { service: 'v1.0.0' },
        configuration_id: 1,
      });
    });

    it('should update configuration when selected', async () => {
      const user = userEvent.setup();
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      );

      const configSelect = screen.getByLabelText('configuration');
      await user.selectOptions(configSelect, '2');

      const submitButton = screen.getByRole('button', { name: /containers:updateVersion/i });
      await user.click(submitButton);

      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          configuration_id: 2,
        })
      );
    });
  });

  describe('compose field validation', () => {
    it('should validate compose field is required', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      const composeInput = screen.getByLabelText(/containers:dockerComposeYaml/i);

      await user.type(versionInput, 'v1.0.0');
      await user.clear(composeInput);

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:composeRequired/i)).toBeInTheDocument();
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('should validate compose max length', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      const composeInput = screen.getByLabelText(/containers:dockerComposeYaml/i);

      await user.type(versionInput, 'v1.0.0');
      await user.clear(composeInput);
      await user.type(composeInput, 'a'.repeat(50001));

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:composeMaxLength/i)).toBeInTheDocument();
    });

    it('should clear validation errors when typing', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const submitButton = screen.getByRole('button', { name: /containers:createVersion/i });
      await user.click(submitButton);

      expect(screen.getByText(/containers:versionRequired/i)).toBeInTheDocument();
      expect(screen.getByText(/containers:composeRequired/i)).toBeInTheDocument();

      const versionInput = screen.getByLabelText(/containers:versionNumber/i);
      await user.type(versionInput, 'v1.0.0');

      expect(screen.queryByText(/containers:versionRequired/i)).not.toBeInTheDocument();

      const composeInput = screen.getByLabelText(/containers:dockerComposeYaml/i);
      await user.type(composeInput, 'version: "3.8"');

      expect(screen.queryByText(/containers:composeRequired/i)).not.toBeInTheDocument();
    });
  });

  describe('loading state', () => {
    it('should disable inputs when loading', () => {
      render(
        <ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} loading={true} />
      );

      expect(screen.getByLabelText(/containers:versionNumber/i)).toBeDisabled();
      expect(screen.getByLabelText(/containers:dockerComposeYaml/i)).toBeDisabled();
      expect(screen.getByRole('button', { name: /common:cancel/i })).toBeDisabled();
      expect(screen.getByRole('button', { name: /containers:creating/i })).toBeDisabled();
    });

    it('should show creating text when loading in create mode', () => {
      render(
        <ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} loading={true} />
      );

      expect(screen.getByRole('button', { name: /containers:creating/i })).toBeInTheDocument();
    });

    it('should show updating text when loading in edit mode', () => {
      render(
        <ContainerVersionForm
          mode="edit"
          initialData={mockInitialVersion}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
          loading={true}
        />
      );

      expect(screen.getByRole('button', { name: /containers:updating/i })).toBeInTheDocument();
    });
  });

  describe('error display', () => {
    it('should display error message when provided', () => {
      render(
        <ContainerVersionForm
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
          error="Failed to create version"
        />
      );

      expect(screen.getByText('Failed to create version')).toBeInTheDocument();
    });

    it('should not display error when null', () => {
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} error={null} />);

      expect(screen.queryByText(/failed/i)).not.toBeInTheDocument();
    });
  });

  describe('cancel action', () => {
    it('should call onCancel when cancel button is clicked', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const cancelButton = screen.getByRole('button', { name: /common:cancel/i });
      await user.click(cancelButton);

      expect(mockOnCancel).toHaveBeenCalled();
    });

    it('should not call onSubmit when cancel is clicked', async () => {
      const user = userEvent.setup();
      render(<ContainerVersionForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);

      const cancelButton = screen.getByRole('button', { name: /common:cancel/i });
      await user.click(cancelButton);

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });
  });
});
