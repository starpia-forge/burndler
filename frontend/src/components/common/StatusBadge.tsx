import React from 'react';
import { ContainerStatus, VersionStatus as ContainerVersionStatus } from '../../types/container';

interface StatusBadgeProps {
  status:
    | ContainerStatus
    | ContainerVersionStatus
    | 'active'
    | 'inactive'
    | 'deleted'
    | 'draft'
    | 'published';
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'outline';
  className?: string;
}

const statusConfig = {
  // Container statuses
  active: {
    label: 'Active',
    bgColor: 'bg-green-100 dark:bg-green-900/20',
    textColor: 'text-green-800 dark:text-green-300',
    borderColor: 'border-green-200 dark:border-green-800',
    outlineBg: 'bg-transparent',
    outlineText: 'text-green-600 dark:text-green-400',
    outlineBorder: 'border-green-300 dark:border-green-600',
  },
  inactive: {
    label: 'Inactive',
    bgColor: 'bg-gray-100 dark:bg-gray-800',
    textColor: 'text-gray-800 dark:text-gray-300',
    borderColor: 'border-gray-200 dark:border-gray-700',
    outlineBg: 'bg-transparent',
    outlineText: 'text-gray-600 dark:text-gray-400',
    outlineBorder: 'border-gray-300 dark:border-gray-600',
  },
  deleted: {
    label: 'Deleted',
    bgColor: 'bg-red-100 dark:bg-red-900/20',
    textColor: 'text-red-800 dark:text-red-300',
    borderColor: 'border-red-200 dark:border-red-800',
    outlineBg: 'bg-transparent',
    outlineText: 'text-red-600 dark:text-red-400',
    outlineBorder: 'border-red-300 dark:border-red-600',
  },
  // Version statuses
  published: {
    label: 'Published',
    bgColor: 'bg-blue-100 dark:bg-blue-900/20',
    textColor: 'text-blue-800 dark:text-blue-300',
    borderColor: 'border-blue-200 dark:border-blue-800',
    outlineBg: 'bg-transparent',
    outlineText: 'text-blue-600 dark:text-blue-400',
    outlineBorder: 'border-blue-300 dark:border-blue-600',
  },
  draft: {
    label: 'Draft',
    bgColor: 'bg-yellow-100 dark:bg-yellow-900/20',
    textColor: 'text-yellow-800 dark:text-yellow-300',
    borderColor: 'border-yellow-200 dark:border-yellow-800',
    outlineBg: 'bg-transparent',
    outlineText: 'text-yellow-600 dark:text-yellow-400',
    outlineBorder: 'border-yellow-300 dark:border-yellow-600',
  },
};

const sizeConfig = {
  sm: {
    padding: 'px-2 py-1',
    text: 'text-xs',
    border: 'border',
  },
  md: {
    padding: 'px-3 py-1',
    text: 'text-sm',
    border: 'border',
  },
  lg: {
    padding: 'px-4 py-2',
    text: 'text-base',
    border: 'border',
  },
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  size = 'sm',
  variant = 'default',
  className = '',
}) => {
  const config = statusConfig[status as keyof typeof statusConfig];
  const sizeStyles = sizeConfig[size];

  if (!config) {
    console.warn(`Unknown status: ${status}`);
    return null;
  }

  const baseClasses = `
    inline-flex items-center justify-center
    rounded-full font-medium
    ${sizeStyles.padding}
    ${sizeStyles.text}
    ${sizeStyles.border}
  `.trim();

  const variantClasses =
    variant === 'outline'
      ? `${config.outlineBg} ${config.outlineText} ${config.outlineBorder}`
      : `${config.bgColor} ${config.textColor} ${config.borderColor}`;

  return <span className={`${baseClasses} ${variantClasses} ${className}`}>{config.label}</span>;
};

// Helper function to get status from container
export const getContainerStatus = (container: {
  active: boolean;
  deleted_at?: string;
}): ContainerStatus => {
  if (container.deleted_at) return ContainerStatus.Deleted;
  return container.active ? ContainerStatus.Active : ContainerStatus.Inactive;
};

// Helper function to get status from container version
export const getContainerVersionStatus = (version: {
  published: boolean;
}): ContainerVersionStatus => {
  return version.published ? ContainerVersionStatus.Published : ContainerVersionStatus.Draft;
};

export default StatusBadge;
