import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import { AuthProvider } from "./providers";

// Self-hosted Geist (latin variable subsets) so the production build has zero
// network dependency on Google Fonts — next/font/google fetches at build time
// and can intermittently fail on Vercel's build machines.
const geistSans = localFont({
  src: "./fonts/GeistSans.woff2",
  variable: "--font-geist-sans",
  weight: "100 900",
});

const geistMono = localFont({
  src: "./fonts/GeistMono.woff2",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "Koola10 CEO Dashboard",
  description: "Mikekoola10Org Autonomous Agent Swarm Control Center",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
