import React, { useEffect, useState } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { ActivityIndicator, View, Text } from 'react-native';
import { useAuthStore } from '@/store/authStore';
import { authAPI } from '@/api/auth';

// Импорты экранов
import LoginScreen from '@/screens/auth/LoginScreen';
import CodeVerificationScreen from '@/screens/auth/CodeVerificationScreen';
import CompleteProfileScreen from '@/screens/auth/CompleteProfileScreen';
import OrganizationsScreen from '@/screens/organizations/OrganizationsScreen';

export type RootStackParamList = {
  Login: undefined;
  CodeVerification: { email: string; isNewUser?: boolean };
  CompleteProfile: undefined;
  Organizations: undefined;
};

const Stack = createStackNavigator<RootStackParamList>();

export default function AppNavigator() {
  const { token, user, isAuthenticated, login } = useAuthStore();
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);

  useEffect(() => {
    checkStoredAuth();
  }, []);

  const checkStoredAuth = async () => {
    try {
      const [storedToken, storedUser] = await Promise.all([
        authAPI.getToken(),
        authAPI.getUser(),
      ]);

      if (storedToken && storedUser) {
        login(storedToken, storedUser);
      }
    } catch (error) {
      console.error('Error checking stored auth:', error);
    } finally {
      setIsCheckingAuth(false);
    }
  };

  if (isCheckingAuth) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator
        screenOptions={{
          headerStyle: { backgroundColor: '#6200ee' },
          headerTintColor: '#fff',
        }}
      >
        {!isAuthenticated ? (
          <>
            <Stack.Screen
              name="Login"
              component={LoginScreen}
              options={{ headerShown: false }}
            />
            <Stack.Screen
              name="CodeVerification"
              component={CodeVerificationScreen}
              options={{ title: 'Ввод кода' }}
            />
          </>
        ) : user?.status === 'pending_profile' ? (
          <Stack.Screen
            name="CompleteProfile"
            component={CompleteProfileScreen}
            options={{ title: 'Заполнение профиля' }}
          />
        ) : (
          <Stack.Screen
            name="Organizations"
            component={OrganizationsScreen}
            options={{ title: 'Организации' }}
          />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}