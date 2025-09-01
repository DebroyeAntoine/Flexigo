use anyhow::{Context, Result};
use std::env;
use std::thread;
use std::time::Duration;
use tts::Tts;

fn estimate_speech_duration(text: &str) -> f64 {
    let word_count = text.split_whitespace().count() as f64;
    let punctuation_count = text.chars().filter(|c| ".,;:!?".contains(*c)).count() as f64;
    // 2.5 words/sec + 300ms per punctuation + margin
    (word_count / 2.5 + punctuation_count * 0.3 + 0.5).max(1.0)
}

fn print_usage() {
    eprintln!("Usage:");
    eprintln!("  tts-rs \"Text to speak\"");
    eprintln!("  tts-rs --voice VoiceName \"Text to speak\"");
    eprintln!("  tts-rs --list-voices");
}

fn main() -> Result<()> {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        print_usage();
        std::process::exit(1);
    }

    let mut tts = Tts::default().context("Failed to initialize TTS")?;

    // Command to list voices
    if args[1] == "--list-voices" {
        if let Ok(voices) = tts.voices() {
            println!("🎵 Available voices:");
            for voice in voices {
                println!("  {} ({})", voice.name(), voice.language());
            }
        }
        return Ok(());
    }

    let (voice_name, text) = if args.len() >= 4 && args[1] == "--voice" {
        // Mode: --voice VoiceName "text"
        (Some(args[2].as_str()), &args[3])
    } else {
        // Simple mode: "text"
        (None, &args[1])
    };

    // Change voice if specified
    if let Some(voice_name) = voice_name {
        if let Ok(voices) = tts.voices() {
            if let Some(voice) = voices.into_iter().find(|v| v.name() == *voice_name) {
                match tts.set_voice(&voice) {
                    Ok(_) => println!("🎭 Voice: {}", voice.name()),
                    Err(e) => eprintln!("⚠️  Failed to change voice: {}", e),
                }
            } else {
                eprintln!("❌ Voice '{}' not found", voice_name);
                eprintln!("💡 Use --list-voices to see available voices");
                std::process::exit(1);
            }
        }
    }

    println!("🔊 Speaking: {}", text);
    tts.speak(text, false)
        .context("Error during speech synthesis")?;

    // Wait for speech to finish
    let duration = estimate_speech_duration(text);
    thread::sleep(Duration::from_millis((duration * 1000.0) as u64));

    // Check if still speaking (with timeout)
    for _ in 0..10 {
        if !tts.is_speaking().unwrap_or(false) {
            break;
        }
        thread::sleep(Duration::from_millis(100));
    }

    println!("✅ Finished");
    Ok(())
}
