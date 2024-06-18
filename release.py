# Given a yaml config file in the prescribed format, generate the required files which can be used to apply the chosen policies on a cluster.
# One file will be generated containing the policies, one for the bindings (these will be created from scratch as per the given yaml), and one for any CRDs (not all policies require these).
# If a policy file begins with '---' as the opening line, this will be stripped.

import yaml
import os
import sys

class Release:
    def __init__(self, config_path, policies_domain, policies_output_filename, bindings_output_filename, crds_output_filename):
        self.__policies_domain = policies_domain
        self.__policies_file = policies_output_filename
        self.__bindings_file = bindings_output_filename
        self.__crds_file = crds_output_filename
        self.__config = self.__parse_config(config_path)

    def __parse_config(self, config_path):
        with open(config_path, 'r') as file:
          config = yaml.safe_load(file)
        return config
    
    def __append_policy(self, subdirectory):
        policy_path = os.path.join('policies', subdirectory, 'policy.yaml')
        if os.path.exists(policy_path):
            with open(policy_path, 'r') as file:
                policy_content = file.read().rstrip('\n').lstrip('---\n')
  
            with open(self.__policies_file, 'a') as file:
                file.write(policy_content + "\n")
                file.write('---\n')
        else:
          print(f"Warning: {policy_path} does not exist.")

    def __create_binding(self, policy_name, binding_name, actions):
        return {
            'apiVersion': 'admissionregistration.k8s.io/v1beta1',
            'kind': 'ValidatingAdmissionPolicyBinding',
            'metadata': {
                'name': binding_name
            },
            'spec': {
                'matchResources': {
                    'matchPolicy': 'Equivalent',
                    'namespaceSelector': {
                        'matchLabels': {f'vap-library.com/{policy_name}': actions[0]}
                    },
                'objectSelector': {}
            },
            'policyName': f'{policy_name}.{self.__policies_domain}',
            'validationActions': actions
            }
        }
    
    def __append_bindings(self, policy_name, bindings):
        for binding in bindings:
          for key in binding:
            actions = binding[key]['validationActions']
            binding_object = self.__create_binding(policy_name, key, actions)
        
          with open(self.__bindings_file, 'a') as file:
              yaml.dump([binding_object], file, default_flow_style=False)
              file.write('---\n')
    
    def __append_crds(self, subdirectory):
        crd_path = os.path.join('policies', subdirectory, 'crd-parameter.yaml')
        if os.path.exists(crd_path):
            with open(crd_path, 'r') as file:
                policy_content = file.read().rstrip('\n').lstrip('---\n')
  
            with open(self.__crds_file, 'a') as file:
                file.write(policy_content + "\n")
                file.write('---\n')
        else:
          print(f"Warning: {crd_path} does not exist.")

    def generate_release_files(self):
        # Ensure the output files are empty at the start
        with open(self.__policies_file, 'w') as file:
            pass
    
        with open(self.__bindings_file, 'w') as file:
            pass
    
        with open(self.__crds_file, 'w') as file:
            pass
    
        for subdirectory, details in self.__config.items():
            # If enabled:true is present we process the entry, else we skip it
            if details.get('enabled', False) == True:
                # Add the requested policy
                self.__append_policy(subdirectory)
                bindings = details.get('bindings', {})
                # Generate the requested bindings
                if bindings:
                    self.__append_bindings(subdirectory, bindings)
                # Add CRDs for the policy if they exist
                self.__append_crds(subdirectory)


def main(config_path):
    policies_file = 'release-policies.yaml'
    bindings_file = 'release-bindings.yaml'
    crds_file = 'release-crds.yaml'
    policies_domain = 'vap-library.com'

    new_release = Release(config_path, policies_domain, policies_file, bindings_file, crds_file)
    new_release.generate_release_files()
        
    print(f"Policies have been appended to {policies_file}")
    print(f"Bindings have been appended to {bindings_file}")
    print(f"CRDs have been appended to {crds_file}, if they exist")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python script.py <config.yaml>")
    else:
        config_path = sys.argv[1]
        main(config_path)
