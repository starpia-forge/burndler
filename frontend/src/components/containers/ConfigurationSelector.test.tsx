import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfigurationSelector } from './ConfigurationSelector';

// Mock hooks
const mockCreateConfig = vi.fn();
vi.mock('../../hooks/useContainerConfigurations', () => ({
  useContainerConfigurations: vi.fn(() => ({
    configurations: [
      { id: 1, name: 'default', minimum_version: 'v1.0.0', description: 'Default config' },
      { id: 2, name: 'production', minimum_version: 'v2.0.0', description: 'Production config' },
    ],
    loading: false,
    createConfig: mockCreateConfig,
  })),
}));

// Mock version validation utility
vi.mock('../../utils/versionCompatibility', () => ({
  isValidSemanticVersion: (version: string) => /^v?\d+\.\d+\.\d+$/.test(version),
}));

describe('ConfigurationSelector', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render dropdown with existing configurations', () => {
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      expect(screen.getByLabelText(/configuration template/i)).toBeInTheDocument();
      expect(screen.getByRole('combobox')).toBeInTheDocument();
      expect(screen.getByText(/no configuration/i)).toBeInTheDocument();
      expect(screen.getByText(/default \(≥ v1\.0\.0\)/i)).toBeInTheDocument();
      expect(screen.getByText(/production \(≥ v2\.0\.0\)/i)).toBeInTheDocument();
      expect(screen.getByText(/\+ create new configuration/i)).toBeInTheDocument();
    });

    it('should show selected configuration', () => {
      render(<ConfigurationSelector containerId="1" value={1} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox') as HTMLSelectElement;
      expect(select.value).toBe('1');
    });

    it('should disable selector when disabled prop is true', () => {
      render(
        <ConfigurationSelector
          containerId="1"
          value={null}
          onChange={mockOnChange}
          disabled={true}
        />
      );

      expect(screen.getByRole('combobox')).toBeDisabled();
    });
  });

  describe('configuration selection', () => {
    it('should call onChange when selecting existing configuration', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.selectOptions(select, '1');

      expect(mockOnChange).toHaveBeenCalledWith(1);
    });

    it('should call onChange with null when selecting no configuration', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={1} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.selectOptions(select, '');

      expect(mockOnChange).toHaveBeenCalledWith(null);
    });

    it('should show create form when selecting create new option', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.selectOptions(select, 'new');

      expect(
        screen.getByRole('heading', { name: /create new configuration/i })
      ).toBeInTheDocument();
      expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/minimum version/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    });
  });

  describe('inline create form', () => {
    it('should validate name field', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      // Open create form
      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(screen.getByText(/name is required/i)).toBeInTheDocument();
    });

    it('should validate name format', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      await user.type(nameInput, 'invalid name!');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(
        screen.getByText(/can only contain letters, numbers, hyphens, and underscores/i)
      ).toBeInTheDocument();
    });

    it('should check for duplicate names', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      await user.type(nameInput, 'default'); // existing name

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(screen.getByText(/configuration with this name already exists/i)).toBeInTheDocument();
    });

    it('should validate minimum version format', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      const versionInput = screen.getByLabelText(/minimum version/i);

      await user.type(nameInput, 'test-config');
      await user.clear(versionInput);
      await user.type(versionInput, 'invalid');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(screen.getByText(/invalid semantic version format/i)).toBeInTheDocument();
    });

    it('should create configuration and call onChange with new ID', async () => {
      const user = userEvent.setup();
      mockCreateConfig.mockResolvedValue({ id: 3, name: 'test-config', minimum_version: 'v1.0.0' });

      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      const versionInput = screen.getByLabelText(/minimum version/i);
      const descInput = screen.getByLabelText(/description/i);

      await user.type(nameInput, 'test-config');
      await user.clear(versionInput);
      await user.type(versionInput, 'v1.2.3');
      await user.type(descInput, 'Test configuration');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      await waitFor(() => {
        expect(mockCreateConfig).toHaveBeenCalledWith({
          name: 'test-config',
          minimum_version: 'v1.2.3',
          description: 'Test configuration',
        });
      });

      await waitFor(() => {
        expect(mockOnChange).toHaveBeenCalledWith(3);
      });
    });

    it('should show loading state during creation', async () => {
      const user = userEvent.setup();
      mockCreateConfig.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 1000))
      );

      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      await user.type(nameInput, 'test-config');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(screen.getByText(/creating/i)).toBeInTheDocument();
      expect(createButton).toBeDisabled();
    });

    it('should close form when clicking close button', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');
      expect(
        screen.getByRole('heading', { name: /create new configuration/i })
      ).toBeInTheDocument();

      const closeButton = screen.getByRole('button', { name: '' }); // X button has no label
      await user.click(closeButton);

      expect(
        screen.queryByRole('heading', { name: /create new configuration/i })
      ).not.toBeInTheDocument();
    });

    it('should show error message on creation failure', async () => {
      const user = userEvent.setup();
      mockCreateConfig.mockRejectedValue(new Error('Creation failed'));

      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const nameInput = screen.getByLabelText(/name/i);
      await user.type(nameInput, 'test-config');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/creation failed/i)).toBeInTheDocument();
      });
    });

    it('should clear errors when typing in fields', async () => {
      const user = userEvent.setup();
      render(<ConfigurationSelector containerId="1" value={null} onChange={mockOnChange} />);

      await user.selectOptions(screen.getByRole('combobox'), 'new');

      const createButton = screen.getByRole('button', { name: /create configuration/i });
      await user.click(createButton);

      expect(screen.getByText(/name is required/i)).toBeInTheDocument();

      const nameInput = screen.getByLabelText(/name/i);
      await user.type(nameInput, 'test');

      expect(screen.queryByText(/name is required/i)).not.toBeInTheDocument();
    });
  });
});
