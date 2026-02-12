/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: {
                    DEFAULT: '#000000',
                    hover: '#1a1a1a',
                },
                secondary: {
                    DEFAULT: '#9CA3AF',
                    hover: '#6B7280',
                },
                background: '#FFFFFF',
                'gray-light': '#F3F4F6',
                'gray-border': '#E5E7EB',
            },
            borderRadius: {
                'button': '12px',
                'input': '12px',
                'card': '16px',
            },
            fontFamily: {
                sans: ['-apple-system', 'BlinkMacSystemFont', 'SF Pro Display', 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', 'sans-serif'],
            },
        },
    },
    plugins: [],
}
