import json
import os


def transform_manifest():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    input_file = os.path.join(script_dir, "q1manifest.json")
    output_file = os.path.join(script_dir, "manifest.json")
    
    animations_folder = os.path.join(script_dir, "standart")
    
    BASE_URL = "http://localhost:9000/public/emoji/standart"
    
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"❌ Error: File {input_file} not found!")
        return
    except json.JSONDecodeError:
        print(f"❌ Error: File {input_file} contains invalid JSON!")
        return
    
    new_manifest = []
    total_emojis = 0
    with_animation = 0
    without_animation = 0
    
    for category in data:
        new_category = {
            "name": category.get('name', ''),
            "slug": category.get('slug', ''),
            "emojis": []
        }
        
        for emoji in category.get('emojis', []):
            emoji_char = emoji.get('emoji', '')
            code = get_unicode_code(emoji_char)
            
            filtered_emoji = {
                "emoji": emoji_char,
                "code": code,
                "skin_tone_support": emoji.get('skin_tone_support', False),
                "name": emoji.get('name', ''),
                "slug": emoji.get('slug', '')
            }
            
            animation_file = os.path.join(animations_folder, f"{code}.json")
            
            if os.path.exists(animation_file):
                filtered_emoji["lottie_url"] = f"{BASE_URL}/{code}.json"
                with_animation += 1
            else:
                filtered_emoji["lottie_url"] = None
                without_animation += 1
            
            new_category['emojis'].append(filtered_emoji)
            total_emojis += 1
        
        new_manifest.append(new_category)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(new_manifest, f, ensure_ascii=False, indent=2)
    
    print(f"✅ Ready!")
    print(f"📊 Categories: {len(new_manifest)}")
    print(f"📊 Total emojis: {total_emojis}")
    print(f"✅ With animation: {with_animation}")
    print(f"❌ Without animation: {without_animation}")

def get_unicode_code(emoji):
    if not emoji:
        return ''
    try:
        return hex(ord(emoji))[2:].upper()
    except:
        return ''

if __name__ == "__main__":
    transform_manifest()
