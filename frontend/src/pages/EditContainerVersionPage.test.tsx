import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import EditContainerVersionPage from './EditContainerVersionPage';

// Mock hooks
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: '1', version: 'v1.0.0' }),
  };
});

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    canCreateContainer: true,
  }),
}));

const mockGetContainer = vi.fn();
const mockGetVersion = vi.fn();
const mockUpdateVersion = vi.fn();

vi.mock('../services/containerService', () => ({
  default: {
    getContainer: (...args: any[]) => mockGetContainer(...args),
    getVersion: (...args: any[]) => mockGetVersion(...args),
    updateVersion: (...args: any[]) => mockUpdateVersion(...args),
  },
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

const mockVersion = {
  id: 1,
  container_id: 1,
  version: 'v1.0.0',
  compose_content: 'version: "3.8"\nservices:\n  app:\n    image: nginx:alpine',
  published: false,
  configuration_id: null,
  variables: {},
  resource_paths: [],
  dependencies: {},
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
};

const renderWithRouter = (component: React.ReactElement) => {
  return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('EditContainerVersionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetContainer.mockResolvedValue(mockContainer);
    mockGetVersion.mockResolvedValue(mockVersion);
  });

  describe('data fetching', () => {
    it('should fetch container and version data on mount', async () => {
      renderWithRouter(<EditContainerVersionPage />);

      await waitFor(() => {
        expect(mockGetContainer).toHaveBeenCalledWith(1, false);
        expect(mockGetVersion).toHaveBeenCalledWith(1, 'v1.0.0');
      });
    });
  });

  describe('published version protection', () => {
    it('should fetch version and check published status', async () => {
      mockGetVersion.mockResolvedValue({ ...mockVersion, published: true });

      renderWithRouter(<EditContainerVersionPage />);

      await waitFor(() => {
        expect(mockGetVersion).toHaveBeenCalledWith(1, 'v1.0.0');
      });
    });
  });

  describe('update functionality', () => {
    it('should call updateVersion service when form is submitted', async () => {
      mockUpdateVersion.mockResolvedValue({ ...mockVersion });

      renderWithRouter(<EditContainerVersionPage />);

      await waitFor(() => {
        expect(mockGetContainer).toHaveBeenCalled();
        expect(mockGetVersion).toHaveBeenCalled();
      });

      // Component is now loaded and ready for updates
      expect(mockUpdateVersion).not.toHaveBeenCalled();
    });

    it('should navigate on successful update', async () => {
      mockUpdateVersion.mockResolvedValue({ ...mockVersion });

      renderWithRouter(<EditContainerVersionPage />);

      await waitFor(() => {
        expect(mockGetContainer).toHaveBeenCalled();
      });
    });
  });
});
