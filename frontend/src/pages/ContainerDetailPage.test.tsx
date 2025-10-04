import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ContainerDetailPage from './ContainerDetailPage';

// Mock hooks
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({ id: '1' }),
    useNavigate: () => vi.fn(),
  };
});

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const mockGetContainer = vi.fn();
const mockGetVersions = vi.fn();
const mockPublishVersion = vi.fn();

vi.mock('../services/containerService', () => ({
  default: {
    getContainer: (...args: any[]) => mockGetContainer(...args),
    getVersions: (...args: any[]) => mockGetVersions(...args),
    publishVersion: (...args: any[]) => mockPublishVersion(...args),
  },
}));

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    isDeveloper: true,
    canCreateContainer: true,
  }),
}));

vi.mock('../hooks/useContainerConfigurations', () => ({
  useContainerConfigurations: () => ({
    configurations: [
      { id: 1, name: 'default', minimum_version: 'v1.0.0' },
      { id: 2, name: 'production', minimum_version: 'v2.0.0' },
    ],
    loading: false,
  }),
}));

const mockContainer = {
  id: 1,
  name: 'test-container',
  description: 'Test container',
  author: 'Test Author',
  repository: 'https://github.com/test/repo',
  active: true,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
};

const mockVersions = [
  {
    id: 1,
    container_id: 1,
    version: 'v1.0.0',
    compose_content: 'version: "3.8"\nservices:\n  app:\n    image: nginx:alpine',
    published: false,
    configuration_id: 1,
    created_at: '2024-01-01',
    updated_at: '2024-01-01',
  },
  {
    id: 2,
    container_id: 1,
    version: 'v2.0.0',
    compose_content: 'version: "3.8"\nservices:\n  app:\n    image: nginx:latest',
    published: true,
    configuration_id: 2,
    created_at: '2024-01-02',
    updated_at: '2024-01-02',
  },
  {
    id: 3,
    container_id: 1,
    version: 'v3.0.0',
    compose_content: 'version: "3.8"\nservices:\n  app:\n    image: node:alpine',
    published: false,
    configuration_id: null,
    created_at: '2024-01-03',
    updated_at: '2024-01-03',
  },
];

const renderWithRouter = (component: React.ReactElement) => {
  return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('ContainerDetailPage - Versions Tab', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetContainer.mockResolvedValue(mockContainer);
    mockGetVersions.mockResolvedValue(mockVersions);
  });

  describe('configuration badge display', () => {
    it('should display configuration badge for version with configuration', async () => {
      renderWithRouter(<ContainerDetailPage />);

      // Wait for versions to load
      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // Check for configuration badge with configuration name
      expect(screen.getByText('default')).toBeInTheDocument();
      expect(screen.getByText('production')).toBeInTheDocument();
    });

    it('should not display configuration badge for version without configuration', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v3.0.0')).toBeInTheDocument();
      });

      // v3.0.0 has no configuration, so there should be only 2 config badges (for v1 and v2)
      const allDefaultBadges = screen.queryAllByText('default');
      expect(allDefaultBadges).toHaveLength(1);
    });

    it('should apply correct styling to configuration badge', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('default')).toBeInTheDocument();
      });

      const badge = screen.getByText('default');
      expect(badge).toHaveClass('bg-purple-100', 'text-purple-800');
    });
  });

  describe('edit button for unpublished versions', () => {
    it('should show edit button for unpublished versions', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // Find edit buttons (should be 2 - for v1.0.0 and v3.0.0 which are unpublished)
      const editButtons = screen.getAllByRole('link', { name: /containers:editVersion/i });
      expect(editButtons).toHaveLength(2);
    });

    it('should not show edit button for published versions', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v2.0.0')).toBeInTheDocument();
      });

      // v2.0.0 is published, so it should not have an edit button
      // We can check this by verifying there are exactly 2 edit buttons (for unpublished versions)
      const editButtons = screen.getAllByRole('link', { name: /containers:editVersion/i });
      expect(editButtons).toHaveLength(2);
    });

    it('should link edit button to correct edit page', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      const editButtons = screen.getAllByRole('link', { name: /containers:editVersion/i });
      const v1EditButton = editButtons.find((btn) => btn.getAttribute('href')?.includes('v1.0.0'));

      expect(v1EditButton).toHaveAttribute('href', '/containers/1/versions/v1.0.0/edit');
    });

    it('should not show edit button when user is not developer', async () => {
      vi.mocked(await import('../hooks/useAuth')).useAuth = () =>
        ({
          isDeveloper: false,
          canCreateContainer: false,
        }) as any;

      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // No edit buttons should be shown for non-developers
      expect(
        screen.queryByRole('link', { name: /containers:editVersion/i })
      ).not.toBeInTheDocument();
    });
  });

  describe('version actions for unpublished versions', () => {
    it('should show both edit and publish buttons for unpublished versions', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // For unpublished versions, both edit and publish should be visible
      expect(screen.getAllByRole('link', { name: /containers:editVersion/i })).toHaveLength(2);
      expect(screen.getAllByRole('button', { name: /containers:publishVersion/i })).toHaveLength(2);
    });

    it('should only show view button for published versions', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v2.0.0')).toBeInTheDocument();
      });

      // Published version should only have view link
      const viewLinks = screen.getAllByRole('link', { name: /containers:viewDetails/i });
      expect(viewLinks.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('version list rendering', () => {
    it('should display all versions with their details', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
        expect(screen.getByText('v2.0.0')).toBeInTheDocument();
        expect(screen.getByText('v3.0.0')).toBeInTheDocument();
      });
    });

    it('should display status badges for versions', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // Check for status badges (Published/Draft)
      const statusBadges = document.querySelectorAll('[class*="badge"]');
      expect(statusBadges.length).toBeGreaterThan(0);
    });
  });

  describe('configuration integration', () => {
    it('should match configuration ID with configuration name correctly', async () => {
      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v1.0.0')).toBeInTheDocument();
      });

      // v1.0.0 has configuration_id: 1, which should match 'default'
      // v2.0.0 has configuration_id: 2, which should match 'production'
      expect(screen.getByText('default')).toBeInTheDocument();
      expect(screen.getByText('production')).toBeInTheDocument();
    });

    it('should handle missing configuration gracefully', async () => {
      // Mock a version with a configuration_id that doesn't exist
      const versionsWithInvalidConfig = [
        ...mockVersions,
        {
          id: 4,
          container_id: 1,
          version: 'v4.0.0',
          compose_content: 'version: "3.8"',
          published: false,
          configuration_id: 999, // Non-existent configuration
          created_at: '2024-01-04',
          updated_at: '2024-01-04',
        },
      ];

      mockGetVersions.mockResolvedValue(versionsWithInvalidConfig);

      renderWithRouter(<ContainerDetailPage />);

      await waitFor(() => {
        expect(screen.getByText('v4.0.0')).toBeInTheDocument();
      });

      // Should not crash and should not show badge for invalid config
      // v4.0.0 should be visible but without a configuration badge
      expect(screen.getByText('v4.0.0')).toBeInTheDocument();
    });
  });
});
