/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Dark mode palette (less OLED)
        ink: {
          950: '#182130',   // deepest background
          900: '#1E2A3D',   // primary background
          850: '#26344A',   // card surface
          800: '#2E3D55',   // elevated surface
          700: '#405269',   // border / divider
          600: '#54677D',   // muted text
          500: '#6B7D93',   // secondary text
          400: '#8B9DAE',   // tertiary text
          300: '#B0C0D0',   // light text
          200: '#D0DAE5',   // bright text
          100: '#ECF0F5',   // brightest text
        },
        // Brand: Mikrotik management - network/ops
        brand: {
          50:  '#EFF6FF',
          400: '#60A5FA',
          500: '#3B82F6',   // primary action
          600: '#2563EB',   // primary
          700: '#1D4ED8',
        },
        // Accent: success / online status
        accent: {
          400: '#34D399',
          500: '#10B981',
          600: '#059669',   // online / success
          700: '#047857',
        },
        // Danger: offline / destructive
        danger: {
          400: '#F87171',
          500: '#EF4444',
          600: '#DC2626',
        },
        // Warning
        warn: {
          400: '#FBBF24',
          500: '#F59E0B',
        },
      },
      fontFamily: {
        sans: ['"Fira Sans"', 'system-ui', 'sans-serif'],
        mono: ['"Fira Code"', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        // Minimal glow for dark mode
        'glow-sm': '0 0 10px rgba(59, 130, 246, 0.15)',
        'glow':    '0 0 20px rgba(59, 130, 246, 0.2)',
        'glow-accent': '0 0 20px rgba(16, 185, 129, 0.2)',
        // Card elevation in dark mode
        'card':    '0 1px 3px rgba(0,0,0,0.4), 0 1px 2px rgba(0,0,0,0.3)',
        'card-hover': '0 4px 12px rgba(0,0,0,0.5), 0 2px 4px rgba(0,0,0,0.4)',
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-out',
        'slide-up': 'slideUp 0.25s ease-out',
        'pulse-glow': 'pulseGlow 2s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        pulseGlow: {
          '0%, 100%': { boxShadow: '0 0 8px rgba(16, 185, 129, 0.3)' },
          '50%': { boxShadow: '0 0 16px rgba(16, 185, 129, 0.5)' },
        },
      },
    },
  },
  plugins: [],
}
