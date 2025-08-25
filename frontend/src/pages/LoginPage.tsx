import React from 'react';
import { Container, Typography, Box } from '@mui/material';

const LoginPage: React.FC = () => {
  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, textAlign: 'center' }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Login Page
        </Typography>
        <Typography variant="body1">
          Login functionality coming soon...
        </Typography>
      </Box>
    </Container>
  );
};

export default LoginPage;
