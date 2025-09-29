import { useTranslation } from 'react-i18next';

interface PlaceholderPageProps {
  titleKey: string;
  descriptionKey: string;
}

export default function PlaceholderPage({ titleKey, descriptionKey }: PlaceholderPageProps) {
  const { t } = useTranslation(['common']);

  return (
    <div className="p-6 bg-card rounded-lg shadow border border-border">
      <h2 className="text-2xl font-bold mb-4 text-foreground">{t(titleKey)}</h2>
      <p className="text-muted-foreground">{t(descriptionKey)}</p>
    </div>
  );
}
