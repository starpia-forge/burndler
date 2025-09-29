import { describe, it, expect } from 'vitest';
import enTranslations from '../en/containers.json';
import koTranslations from '../ko/containers.json';

describe('Translation Files', () => {
  describe('JSON Syntax Validation', () => {
    it('should have valid JSON syntax for English translations', () => {
      expect(enTranslations).toBeDefined();
      expect(typeof enTranslations).toBe('object');
    });

    it('should have valid JSON syntax for Korean translations', () => {
      expect(koTranslations).toBeDefined();
      expect(typeof koTranslations).toBe('object');
    });
  });

  describe('Translation Key Consistency', () => {
    it('should have the same keys in both English and Korean translations', () => {
      const enKeys = Object.keys(enTranslations).sort();
      const koKeys = Object.keys(koTranslations).sort();

      expect(enKeys).toEqual(koKeys);
    });

    it('should not have empty translation values', () => {
      const enValues = Object.values(enTranslations);
      const koValues = Object.values(koTranslations);

      enValues.forEach((value, _index) => {
        expect(value.trim()).not.toBe('');
      });

      koValues.forEach((value, _index) => {
        expect(value.trim()).not.toBe('');
      });
    });
  });

  describe('Interpolation Syntax Validation', () => {
    it('should use correct interpolation syntax (double curly braces) in English', () => {
      // This regex matches single curly braces that are NOT part of double curly braces
      const singleBracePattern = /\{(?!\{)[^}]*\}(?!\})/g;

      Object.entries(enTranslations).forEach(([key, value]) => {
        const matches = value.match(singleBracePattern);
        if (matches) {
          console.warn(
            `Found single curly brace interpolation in English key "${key}": ${matches.join(', ')}`
          );
          expect(matches).toBeNull();
        }
      });
    });

    it('should use correct interpolation syntax (double curly braces) in Korean', () => {
      // This regex matches single curly braces that are NOT part of double curly braces
      const singleBracePattern = /\{(?!\{)[^}]*\}(?!\})/g;

      Object.entries(koTranslations).forEach(([key, value]) => {
        const matches = value.match(singleBracePattern);
        if (matches) {
          console.warn(
            `Found single curly brace interpolation in Korean key "${key}": ${matches.join(', ')}`
          );
          expect(matches).toBeNull();
        }
      });
    });

    it('should have core interpolation variables matching between languages', () => {
      const extractVariables = (text: string): string[] => {
        const matches = text.match(/\{\{(\w+)\}\}/g);
        return matches ? matches.map((match) => match.slice(2, -2)) : [];
      };

      Object.keys(enTranslations).forEach((key) => {
        const enVariables = extractVariables(enTranslations[key as keyof typeof enTranslations]);
        const koVariables = extractVariables(koTranslations[key as keyof typeof koTranslations]);

        // Filter out language-specific variables like 's' for pluralization
        const coreEnVariables = enVariables.filter((v) => v !== 's').sort();
        const coreKoVariables = koVariables.filter((v) => v !== 's').sort();

        if (coreEnVariables.length > 0 || coreKoVariables.length > 0) {
          expect(coreEnVariables).toEqual(coreKoVariables);
        }
      });
    });
  });

  describe('Translation Quality', () => {
    it('should not contain HTML tags in translation values', () => {
      const htmlTagPattern = /<[^>]*>/;

      Object.entries(enTranslations).forEach(([_key, value]) => {
        expect(htmlTagPattern.test(value)).toBe(false);
      });

      Object.entries(koTranslations).forEach(([_key, value]) => {
        expect(htmlTagPattern.test(value)).toBe(false);
      });
    });

    it('should not have excessive duplicate translation values in same language', () => {
      const enValues = Object.values(enTranslations);
      const koValues = Object.values(koTranslations);

      // Allow common words to be duplicated
      const allowedDuplicates = [
        'Edit',
        'Delete',
        'View',
        'Create',
        'Update',
        'Cancel',
        'Confirm',
        '편집',
        '삭제',
        '보기',
        '생성',
        '수정',
        '취소',
        '확인',
        '버전',
      ];

      const enFiltered = enValues.filter((v) => !allowedDuplicates.includes(v));
      const koFiltered = koValues.filter((v) => !allowedDuplicates.includes(v));

      // Check that we don't have more than 90% unique values (some duplication is acceptable)
      const enUniqueRatio = new Set(enFiltered).size / enFiltered.length;
      const koUniqueRatio = new Set(koFiltered).size / koFiltered.length;

      expect(enUniqueRatio).toBeGreaterThan(0.9);
      expect(koUniqueRatio).toBeGreaterThan(0.9);
    });
  });
});
