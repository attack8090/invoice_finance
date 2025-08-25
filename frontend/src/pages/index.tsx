// Placeholder pages for the React app

import React from 'react';
import { Container, Typography, Box } from '@mui/material';

export const RegisterPage: React.FC = () => (
  <Container maxWidth="sm">
    <Box sx={{ mt: 8, textAlign: 'center' }}>
      <Typography variant="h4" component="h1" gutterBottom>Register</Typography>
      <Typography>Registration form coming soon...</Typography>
    </Box>
  </Container>
);

export const SMEDashboard: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>SME Dashboard</Typography>
      <Typography>SME dashboard with invoice management coming soon...</Typography>
    </Box>
  </Container>
);

export const InvestorDashboard: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>Investor Dashboard</Typography>
      <Typography>Investor dashboard with portfolio management coming soon...</Typography>
    </Box>
  </Container>
);

export const AdminDashboard: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>Admin Dashboard</Typography>
      <Typography>Admin dashboard with platform management coming soon...</Typography>
    </Box>
  </Container>
);

export const InvoicesPage: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>Invoices</Typography>
      <Typography>Invoice management page coming soon...</Typography>
    </Box>
  </Container>
);

export const InvestmentsPage: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>Investments</Typography>
      <Typography>Investment portfolio page coming soon...</Typography>
    </Box>
  </Container>
);

export const MarketplacePage: React.FC = () => (
  <Container maxWidth="lg">
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>Marketplace</Typography>
      <Typography>Investment marketplace coming soon...</Typography>
    </Box>
  </Container>
);
