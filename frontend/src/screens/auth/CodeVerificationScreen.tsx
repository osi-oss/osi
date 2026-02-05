// src/screens/auth/CodeVerificationScreen.tsx
import React, { useState } from 'react';
import {
  View,
  StyleSheet,
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { Button, Text, Surface, ActivityIndicator } from 'react-native-paper';
import { useNavigation, useRoute } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import { RootStackParamList } from '@/navigation/AppNavigator';
import { authAPI } from '@/api/auth';
import { useAuthStore } from '@/store/authStore';
import CodeInput from '@/components/ui/CodeInput';

type NavigationProp = StackNavigationProp<RootStackParamList, 'CodeVerification'>;

export default function CodeVerificationScreen() {
  const navigation = useNavigation<NavigationProp>();
  const route = useRoute();
  const { email, isNewUser } = route.params as { email: string; isNewUser?: boolean };
  
  const { login } = useAuthStore();
  const [code, setCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [timer, setTimer] = useState(120); // 2 минуты в секундах

  const handleVerifyCode = async () => {
    if (code.length !== 4) {
      Alert.alert('Ошибка', 'Введите 4-значный код');
      return;
    }

    try {
      setIsLoading(true);
      
      const response = await authAPI.verifyCode(email, code);
      
      // Сохраняем токен и пользователя
      await authAPI.saveToken(response.token);
      await authAPI.saveUser(response.user);
      
      // Обновляем store
      login(response.token, response.user);

      // Редирект
      if (response.user.status === 'pending_profile') {
        navigation.navigate('CompleteProfile');
      } else {
        navigation.navigate('Organizations');
      }

    } catch (error: any) {
      Alert.alert(
        'Ошибка',
        error.error || 'Неверный код. Попробуйте снова.',
        [{ text: 'OK' }]
      );
      setCode('');
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendCode = async () => {
    try {
      setIsLoading(true);
      await authAPI.requestCode(email);
      Alert.alert('Успешно', 'Новый код отправлен на вашу почту');
      setTimer(120); // Сброс таймера
    } catch (error: any) {
      Alert.alert('Ошибка', error.error || 'Не удалось отправить код');
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
        <Surface style={styles.content} elevation={2}>
          <Text style={styles.title}>Введите код</Text>
          
          <Text style={styles.subtitle}>
            Мы отправили код на{'\n'}
            <Text style={styles.email}>{email}</Text>
          </Text>

          <View style={styles.codeContainer}>
            <CodeInput
              length={4}
              onCodeChange={setCode}
              autoFocus={true}
            />
          </View>

          <Button
            mode="contained"
            onPress={handleVerifyCode}
            loading={isLoading}
            disabled={isLoading || code.length !== 4}
            style={styles.button}
          >
            Подтвердить
          </Button>

          <View style={styles.timerContainer}>
            <Text style={styles.timerText}>
              Отправить новый код можно через {Math.floor(timer / 60)}:
              {(timer % 60).toString().padStart(2, '0')}
            </Text>
          </View>

          <Button
            mode="outlined"
            onPress={handleResendCode}
            disabled={timer > 0 || isLoading}
            style={styles.resendButton}
          >
            Отправить код снова
          </Button>

          <Button
            mode="text"
            onPress={() => navigation.goBack()}
            disabled={isLoading}
            style={styles.backButton}
          >
            Изменить email
          </Button>
        </Surface>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
  },
  content: {
    margin: 20,
    padding: 24,
    borderRadius: 16,
    backgroundColor: 'white',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 32,
    color: '#666',
    lineHeight: 22,
  },
  email: {
    fontWeight: '600',
    color: '#6200ee',
  },
  codeContainer: {
    marginBottom: 24,
  },
  button: {
    marginBottom: 16,
  },
  timerContainer: {
    alignItems: 'center',
    marginBottom: 16,
  },
  timerText: {
    fontSize: 14,
    color: '#666',
  },
  resendButton: {
    marginBottom: 8,
  },
  backButton: {
    marginTop: 8,
  },
});