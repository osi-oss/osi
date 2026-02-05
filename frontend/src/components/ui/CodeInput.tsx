// src/components/ui/CodeInput.tsx
import React, { useState, useRef } from 'react';
import { View, TextInput, StyleSheet, Keyboard } from 'react-native';

interface CodeInputProps {
  length?: number;
  onCodeChange: (code: string) => void;
  autoFocus?: boolean;
}

export default function CodeInput({ 
  length = 4, 
  onCodeChange, 
  autoFocus = true 
}: CodeInputProps) {
  const [code, setCode] = useState<string[]>(new Array(length).fill(''));
  const inputs = useRef<TextInput[]>([]);

  const handleChange = (text: string, index: number) => {
    const newCode = [...code];
    newCode[index] = text;
    setCode(newCode);

    const codeString = newCode.join('');
    onCodeChange(codeString);

    // Автопереход к следующему полю
    if (text && index < length - 1) {
      inputs.current[index + 1]?.focus();
    }

    // Если все поля заполнены, скрыть клавиатуру
    if (codeString.length === length) {
      Keyboard.dismiss();
    }
  };

  const handleKeyPress = (e: any, index: number) => {
    if (e.nativeEvent.key === 'Backspace') {
      if (!code[index] && index > 0) {
        inputs.current[index - 1]?.focus();
      }
    }
  };

  return (
    <View style={styles.container}>
      {Array.from({ length }).map((_, index) => (
        <TextInput
          key={index}
          ref={(ref) => {
            if (ref) inputs.current[index] = ref;
          }}
          style={[styles.input, code[index] ? styles.filled : styles.empty]}
          value={code[index]}
          onChangeText={(text) => handleChange(text, index)}
          onKeyPress={(e) => handleKeyPress(e, index)}
          maxLength={1}
          keyboardType="number-pad"
          autoFocus={autoFocus && index === 0}
          selectTextOnFocus
        />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 10,
  },
  input: {
    width: 50,
    height: 50,
    borderWidth: 2,
    borderRadius: 10,
    textAlign: 'center',
    fontSize: 24,
    fontWeight: 'bold',
  },
  empty: {
    borderColor: '#ddd',
    backgroundColor: '#f9f9f9',
  },
  filled: {
    borderColor: '#6200ee',
    backgroundColor: '#f3e5f5',
  },
});