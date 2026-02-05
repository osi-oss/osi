// src/screens/auth/LoginScreen.tsx
import React, { useState } from 'react';
import {
  View,
  StyleSheet,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import { Button, TextInput, Text, Surface } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { authAPI, RequestCodeResponse } from '@/api/auth'; // Импортируйте тип
import { RootStackParamList } from '@/navigation/AppNavigator';

const schema = yup.object({
  email: yup
    .string()
    .email('Введите корректный email адрес')
    .required('Email обязателен для заполнения'),
});

type FormData = yup.InferType<typeof schema>;
type NavigationProp = StackNavigationProp<RootStackParamList, 'Login'>;

export default function LoginScreen() {
  const navigation = useNavigation<NavigationProp>();
  const [isLoading, setIsLoading] = useState(false);

  const {
    control,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<FormData>({
    resolver: yupResolver(schema),
    defaultValues: { email: '' },
  });

  const email = watch('email');

  const handleRequestCode = async (data: FormData) => {
    try {
      setIsLoading(true);
      
      // Типизированный вызов
      const response = await authAPI.requestCode(data.email) as unknown as RequestCodeResponse;
      
      navigation.navigate('CodeVerification', {
        email: data.email,
        isNewUser: response.is_new_user,
      });
      
    } catch (error: any) {
      console.error('Request code error:', error);
      
      Alert.alert(
        'Ошибка',
        error.error || 'Не удалось отправить код. Проверьте email и попробуйте снова.',
        [{ text: 'OK' }]
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scrollContent}>
        <View style={styles.content}>
          <Surface style={styles.logoContainer} elevation={2}>
            <Text style={styles.logoText}>OSI</Text>
            <Text style={styles.logoSubtitle}>Organization System</Text>
          </Surface>

          <Text style={styles.title}>Вход в систему</Text>
          
          <Controller
            control={control}
            name="email"
            render={({ field: { onChange, value } }) => (
              <TextInput
                label="Email"
                value={value}
                onChangeText={onChange}
                mode="outlined"
                style={styles.input}
                error={!!errors.email}
                left={<TextInput.Icon icon="email" />}
                disabled={isLoading}
              />
            )}
          />
          
          {errors.email && (
            <Text style={styles.errorText}>{errors.email.message}</Text>
          )}

          <Button
            mode="contained"
            onPress={handleSubmit(handleRequestCode)}
            loading={isLoading}
            disabled={isLoading || !email || !!errors.email}
            style={styles.button}
          >
            Получить код
          </Button>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f5f5f5' },
  scrollContent: { flexGrow: 1, justifyContent: 'center' },
  content: { padding: 24 },
  logoContainer: {
    alignSelf: 'center',
    padding: 20,
    borderRadius: 16,
    marginBottom: 32,
    alignItems: 'center',
    backgroundColor: '#6200ee',
  },
  logoText: { fontSize: 48, fontWeight: 'bold', color: 'white' },
  logoSubtitle: { fontSize: 14, color: 'white', opacity: 0.8, marginTop: 4 },
  title: { fontSize: 28, fontWeight: 'bold', textAlign: 'center', marginBottom: 8 },
  input: { marginBottom: 8, backgroundColor: 'white' },
  errorText: { color: '#d32f2f', fontSize: 14, marginBottom: 16 },
  button: { marginTop: 8 },
});