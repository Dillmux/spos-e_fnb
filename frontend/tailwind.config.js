/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        'brand-primary': '#30AFFF',
        'brand-cyan': '#92EEFF',
        'brand-mint-light': '#D8FFC5',
        'brand-mint-active': '#C4F7CA',
        'brand-dark': '#171C21',
        'brand-surface': '#F7F9FF',
        'brand-danger': '#BA1A1A',
        surface: '#F7F9FF',
        'surface-dim': '#D6DAE1',
        'surface-container-lowest': '#FFFFFF',
        'surface-container-low': '#F0F4FB',
        'surface-container': '#EAEEF5',
        'surface-container-high': '#E4E8EF',
        'surface-container-highest': '#DFE3E9',
        'on-surface': '#171C21',
        'on-surface-variant': '#3F4851',
        outline: '#6F7882',
        'outline-variant': '#BEC7D3',
        primary: '#006397',
        'primary-container': '#30AFFF',
        'on-primary-container': '#004063',
        secondary: '#006875',
        'secondary-container': '#92EEFF',
        tertiary: '#46673A',
        'tertiary-container': '#8DB17D',
        'tertiary-fixed': '#C7EDB5',
        error: '#BA1A1A',
        'error-container': '#FFDAD6',
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'sans-serif'],
      },
      minHeight: {
        touch: '48px',
      },
      minWidth: {
        touch: '48px',
      },
      borderRadius: {
        sm: '0.25rem',
        DEFAULT: '0.5rem',
        md: '0.75rem',
        lg: '1rem',
        xl: '1.5rem',
      },
    },
  },
  plugins: [],
};

